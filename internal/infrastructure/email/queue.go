package email

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	// Queue keys
	emailQueueKey       = "email:queue"
	emailProcessingKey  = "email:processing"
	emailDeadLetterKey  = "email:dead_letter"
	emailScheduledKey   = "email:scheduled"
	emailStatsKey       = "email:stats"

	// Retry settings
	maxRetries          = 3
	retryDelayBase      = time.Minute
	processingTimeout   = 5 * time.Minute
)

// EmailQueue manages email queuing and processing
type EmailQueue struct {
	redis  *redis.Client
	config *Config
}

// NewEmailQueue creates a new email queue
func NewEmailQueue(redisClient *redis.Client, config *Config) *EmailQueue {
	return &EmailQueue{
		redis:  redisClient,
		config: config,
	}
}

// Enqueue adds an email to the queue
func (q *EmailQueue) Enqueue(ctx context.Context, email *Email) error {
	email.Status = EmailStatusPending
	email.CreatedAt = time.Now()

	data, err := json.Marshal(email)
	if err != nil {
		return fmt.Errorf("failed to marshal email: %w", err)
	}

	// If scheduled, add to scheduled set
	if email.ScheduledAt != nil && email.ScheduledAt.After(time.Now()) {
		score := float64(email.ScheduledAt.Unix())
		return q.redis.ZAdd(ctx, emailScheduledKey, redis.Z{
			Score:  score,
			Member: data,
		}).Err()
	}

	// Otherwise, add to immediate queue
	return q.redis.LPush(ctx, emailQueueKey, data).Err()
}

// EnqueueWithDelay adds an email to the queue with a delay
func (q *EmailQueue) EnqueueWithDelay(ctx context.Context, email *Email, delay time.Duration) error {
	scheduledAt := time.Now().Add(delay)
	email.ScheduledAt = &scheduledAt
	return q.Enqueue(ctx, email)
}

// Dequeue gets the next email from the queue
func (q *EmailQueue) Dequeue(ctx context.Context) (*Email, error) {
	// First, check for scheduled emails that are due
	if err := q.processScheduledEmails(ctx); err != nil {
		// Log but don't fail
		fmt.Printf("Error processing scheduled emails: %v\n", err)
	}

	// Move item from queue to processing with timeout
	result, err := q.redis.BRPopLPush(ctx, emailQueueKey, emailProcessingKey, 5*time.Second).Result()
	if err == redis.Nil {
		return nil, nil // No items in queue
	}
	if err != nil {
		return nil, fmt.Errorf("failed to dequeue: %w", err)
	}

	var email Email
	if err := json.Unmarshal([]byte(result), &email); err != nil {
		return nil, fmt.Errorf("failed to unmarshal email: %w", err)
	}

	return &email, nil
}

// processScheduledEmails moves due scheduled emails to the main queue
func (q *EmailQueue) processScheduledEmails(ctx context.Context) error {
	now := float64(time.Now().Unix())

	// Get all scheduled emails that are due
	results, err := q.redis.ZRangeByScore(ctx, emailScheduledKey, &redis.ZRangeBy{
		Min: "-inf",
		Max: fmt.Sprintf("%f", now),
	}).Result()
	if err != nil {
		return err
	}

	for _, result := range results {
		// Move to main queue
		if err := q.redis.LPush(ctx, emailQueueKey, result).Err(); err != nil {
			continue
		}
		// Remove from scheduled
		q.redis.ZRem(ctx, emailScheduledKey, result)
	}

	return nil
}

// Complete marks an email as successfully sent
func (q *EmailQueue) Complete(ctx context.Context, email *Email) error {
	data, err := json.Marshal(email)
	if err != nil {
		return err
	}

	// Remove from processing
	q.redis.LRem(ctx, emailProcessingKey, 1, data)

	// Update stats
	q.incrementStat(ctx, "sent")
	q.incrementStat(ctx, fmt.Sprintf("sent:%s", email.Type))

	return nil
}

// Fail marks an email as failed and potentially retries
func (q *EmailQueue) Fail(ctx context.Context, email *Email, err error) error {
	email.RetryCount++
	email.Error = err.Error()
	email.Status = EmailStatusFailed

	data, _ := json.Marshal(email)

	// Remove from processing
	q.redis.LRem(ctx, emailProcessingKey, 1, data)

	if email.RetryCount < maxRetries {
		// Re-queue with exponential backoff
		delay := retryDelayBase * time.Duration(1<<uint(email.RetryCount))
		return q.EnqueueWithDelay(ctx, email, delay)
	}

	// Move to dead letter queue
	q.redis.LPush(ctx, emailDeadLetterKey, data)
	q.incrementStat(ctx, "failed")

	return nil
}

// GetQueueLength returns the current queue length
func (q *EmailQueue) GetQueueLength(ctx context.Context) (int64, error) {
	return q.redis.LLen(ctx, emailQueueKey).Result()
}

// GetProcessingLength returns the number of emails being processed
func (q *EmailQueue) GetProcessingLength(ctx context.Context) (int64, error) {
	return q.redis.LLen(ctx, emailProcessingKey).Result()
}

// GetDeadLetterLength returns the number of failed emails
func (q *EmailQueue) GetDeadLetterLength(ctx context.Context) (int64, error) {
	return q.redis.LLen(ctx, emailDeadLetterKey).Result()
}

// GetScheduledLength returns the number of scheduled emails
func (q *EmailQueue) GetScheduledLength(ctx context.Context) (int64, error) {
	return q.redis.ZCard(ctx, emailScheduledKey).Result()
}

// GetStats returns email statistics
func (q *EmailQueue) GetStats(ctx context.Context) (map[string]int64, error) {
	result, err := q.redis.HGetAll(ctx, emailStatsKey).Result()
	if err != nil {
		return nil, err
	}

	stats := make(map[string]int64)
	for k, v := range result {
		var val int64
		fmt.Sscanf(v, "%d", &val)
		stats[k] = val
	}

	return stats, nil
}

// incrementStat increments a stat counter
func (q *EmailQueue) incrementStat(ctx context.Context, stat string) {
	q.redis.HIncrBy(ctx, emailStatsKey, stat, 1)

	// Also increment daily stat
	today := time.Now().Format("2006-01-02")
	q.redis.HIncrBy(ctx, fmt.Sprintf("%s:%s", emailStatsKey, today), stat, 1)
}

// RecoverStaleProcessing recovers emails stuck in processing
func (q *EmailQueue) RecoverStaleProcessing(ctx context.Context) error {
	// Get all items in processing queue
	items, err := q.redis.LRange(ctx, emailProcessingKey, 0, -1).Result()
	if err != nil {
		return err
	}

	for _, item := range items {
		var email Email
		if err := json.Unmarshal([]byte(item), &email); err != nil {
			continue
		}

		// If email has been processing for too long, re-queue it
		if time.Since(email.CreatedAt) > processingTimeout {
			// Remove from processing
			q.redis.LRem(ctx, emailProcessingKey, 1, item)

			// Re-queue
			email.RetryCount++
			if email.RetryCount < maxRetries {
				q.Enqueue(ctx, &email)
			} else {
				// Move to dead letter
				data, _ := json.Marshal(email)
				q.redis.LPush(ctx, emailDeadLetterKey, data)
			}
		}
	}

	return nil
}

// RetryDeadLetter retries a failed email from dead letter queue
func (q *EmailQueue) RetryDeadLetter(ctx context.Context, emailID string) error {
	// Get all items in dead letter queue
	items, err := q.redis.LRange(ctx, emailDeadLetterKey, 0, -1).Result()
	if err != nil {
		return err
	}

	for _, item := range items {
		var email Email
		if err := json.Unmarshal([]byte(item), &email); err != nil {
			continue
		}

		if email.ID == emailID {
			// Remove from dead letter
			q.redis.LRem(ctx, emailDeadLetterKey, 1, item)

			// Reset retry count and re-queue
			email.RetryCount = 0
			email.Error = ""
			email.Status = EmailStatusPending
			return q.Enqueue(ctx, &email)
		}
	}

	return fmt.Errorf("email not found in dead letter queue: %s", emailID)
}

// ClearDeadLetter clears the dead letter queue
func (q *EmailQueue) ClearDeadLetter(ctx context.Context) error {
	return q.redis.Del(ctx, emailDeadLetterKey).Err()
}
