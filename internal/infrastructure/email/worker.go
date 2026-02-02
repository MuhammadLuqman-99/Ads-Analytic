package email

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// Worker processes emails from the queue
type Worker struct {
	queue       *EmailQueue
	service     EmailService
	config      *Config
	concurrency int
	stopCh      chan struct{}
	wg          sync.WaitGroup
	running     bool
	mu          sync.Mutex
}

// WorkerConfig contains worker configuration
type WorkerConfig struct {
	Concurrency       int           // Number of concurrent workers
	PollInterval      time.Duration // How often to poll for new emails
	RecoveryInterval  time.Duration // How often to recover stale emails
}

// DefaultWorkerConfig returns default worker configuration
func DefaultWorkerConfig() *WorkerConfig {
	return &WorkerConfig{
		Concurrency:      5,
		PollInterval:     time.Second,
		RecoveryInterval: 5 * time.Minute,
	}
}

// NewWorker creates a new email worker
func NewWorker(queue *EmailQueue, service EmailService, config *Config, workerConfig *WorkerConfig) *Worker {
	if workerConfig == nil {
		workerConfig = DefaultWorkerConfig()
	}

	return &Worker{
		queue:       queue,
		service:     service,
		config:      config,
		concurrency: workerConfig.Concurrency,
		stopCh:      make(chan struct{}),
	}
}

// Start starts the email worker
func (w *Worker) Start(ctx context.Context) error {
	w.mu.Lock()
	if w.running {
		w.mu.Unlock()
		return fmt.Errorf("worker already running")
	}
	w.running = true
	w.mu.Unlock()

	log.Printf("[EmailWorker] Starting with %d concurrent workers", w.concurrency)

	// Start worker goroutines
	for i := 0; i < w.concurrency; i++ {
		w.wg.Add(1)
		go w.processLoop(ctx, i)
	}

	// Start recovery goroutine
	w.wg.Add(1)
	go w.recoveryLoop(ctx)

	// Start stats reporter
	w.wg.Add(1)
	go w.statsLoop(ctx)

	return nil
}

// Stop stops the email worker
func (w *Worker) Stop() {
	w.mu.Lock()
	if !w.running {
		w.mu.Unlock()
		return
	}
	w.running = false
	w.mu.Unlock()

	log.Println("[EmailWorker] Stopping...")
	close(w.stopCh)
	w.wg.Wait()
	log.Println("[EmailWorker] Stopped")
}

// processLoop continuously processes emails
func (w *Worker) processLoop(ctx context.Context, workerID int) {
	defer w.wg.Done()

	log.Printf("[EmailWorker-%d] Started", workerID)

	for {
		select {
		case <-ctx.Done():
			log.Printf("[EmailWorker-%d] Context cancelled", workerID)
			return
		case <-w.stopCh:
			log.Printf("[EmailWorker-%d] Stop signal received", workerID)
			return
		default:
			if err := w.processOne(ctx, workerID); err != nil {
				log.Printf("[EmailWorker-%d] Error: %v", workerID, err)
				time.Sleep(time.Second) // Backoff on error
			}
		}
	}
}

// processOne processes a single email
func (w *Worker) processOne(ctx context.Context, workerID int) error {
	email, err := w.queue.Dequeue(ctx)
	if err != nil {
		return fmt.Errorf("failed to dequeue: %w", err)
	}

	if email == nil {
		// No emails in queue, wait a bit
		time.Sleep(time.Second)
		return nil
	}

	log.Printf("[EmailWorker-%d] Processing email %s (type: %s, to: %s)",
		workerID, email.ID, email.Type, email.To)

	// Send the email
	startTime := time.Now()
	err = w.service.Send(ctx, email)
	duration := time.Since(startTime)

	if err != nil {
		log.Printf("[EmailWorker-%d] Failed to send email %s: %v", workerID, email.ID, err)
		return w.queue.Fail(ctx, email, err)
	}

	log.Printf("[EmailWorker-%d] Sent email %s in %v", workerID, email.ID, duration)

	now := time.Now()
	email.SentAt = &now
	email.Status = EmailStatusSent

	return w.queue.Complete(ctx, email)
}

// recoveryLoop periodically recovers stale emails
func (w *Worker) recoveryLoop(ctx context.Context) {
	defer w.wg.Done()

	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-w.stopCh:
			return
		case <-ticker.C:
			if err := w.queue.RecoverStaleProcessing(ctx); err != nil {
				log.Printf("[EmailWorker] Recovery error: %v", err)
			}
		}
	}
}

// statsLoop periodically logs stats
func (w *Worker) statsLoop(ctx context.Context) {
	defer w.wg.Done()

	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-w.stopCh:
			return
		case <-ticker.C:
			w.logStats(ctx)
		}
	}
}

// logStats logs current queue statistics
func (w *Worker) logStats(ctx context.Context) {
	queueLen, _ := w.queue.GetQueueLength(ctx)
	processingLen, _ := w.queue.GetProcessingLength(ctx)
	deadLetterLen, _ := w.queue.GetDeadLetterLength(ctx)
	scheduledLen, _ := w.queue.GetScheduledLength(ctx)

	log.Printf("[EmailWorker] Stats: queue=%d, processing=%d, dead_letter=%d, scheduled=%d",
		queueLen, processingLen, deadLetterLen, scheduledLen)
}

// GetStats returns current worker statistics
func (w *Worker) GetStats(ctx context.Context) (*WorkerStats, error) {
	queueLen, err := w.queue.GetQueueLength(ctx)
	if err != nil {
		return nil, err
	}

	processingLen, err := w.queue.GetProcessingLength(ctx)
	if err != nil {
		return nil, err
	}

	deadLetterLen, err := w.queue.GetDeadLetterLength(ctx)
	if err != nil {
		return nil, err
	}

	scheduledLen, err := w.queue.GetScheduledLength(ctx)
	if err != nil {
		return nil, err
	}

	stats, err := w.queue.GetStats(ctx)
	if err != nil {
		return nil, err
	}

	return &WorkerStats{
		QueueLength:      queueLen,
		ProcessingLength: processingLen,
		DeadLetterLength: deadLetterLen,
		ScheduledLength:  scheduledLen,
		TotalSent:        stats["sent"],
		TotalFailed:      stats["failed"],
		Running:          w.running,
		Concurrency:      w.concurrency,
	}, nil
}

// WorkerStats contains worker statistics
type WorkerStats struct {
	QueueLength      int64 `json:"queue_length"`
	ProcessingLength int64 `json:"processing_length"`
	DeadLetterLength int64 `json:"dead_letter_length"`
	ScheduledLength  int64 `json:"scheduled_length"`
	TotalSent        int64 `json:"total_sent"`
	TotalFailed      int64 `json:"total_failed"`
	Running          bool  `json:"running"`
	Concurrency      int   `json:"concurrency"`
}
