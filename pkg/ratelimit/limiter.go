package ratelimit

import (
	"context"
	"sync"
	"time"
)

// Limiter implements a token bucket rate limiter
type Limiter struct {
	mu         sync.Mutex
	tokens     float64
	maxTokens  float64
	refillRate float64 // tokens per second
	lastRefill time.Time
	waiters    []chan struct{}
}

// NewLimiter creates a new rate limiter
// calls: number of allowed calls in the window
// window: the time window for rate limiting
func NewLimiter(calls int, window time.Duration) *Limiter {
	refillRate := float64(calls) / window.Seconds()
	return &Limiter{
		tokens:     float64(calls),
		maxTokens:  float64(calls),
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

// Allow checks if a request is allowed immediately
func (l *Limiter) Allow() bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.refill()

	if l.tokens >= 1 {
		l.tokens--
		return true
	}
	return false
}

// Wait waits until a request is allowed or context is cancelled
func (l *Limiter) Wait(ctx context.Context) error {
	// Try immediate allow first
	if l.Allow() {
		return nil
	}

	// Calculate wait time
	waitTime := l.timeUntilAvailable()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(waitTime):
		// Try again after waiting
		if l.Allow() {
			return nil
		}
		// Recursively wait if still not available
		return l.Wait(ctx)
	}
}

// WaitN waits for n tokens to be available
func (l *Limiter) WaitN(ctx context.Context, n int) error {
	for i := 0; i < n; i++ {
		if err := l.Wait(ctx); err != nil {
			return err
		}
	}
	return nil
}

// Reserve reserves a token and returns the time to wait before using it
func (l *Limiter) Reserve() time.Duration {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.refill()

	if l.tokens >= 1 {
		l.tokens--
		return 0
	}

	// Calculate time until a token is available
	deficit := 1 - l.tokens
	waitTime := time.Duration(deficit/l.refillRate) * time.Second
	l.tokens = 0

	return waitTime
}

// Tokens returns the current number of available tokens
func (l *Limiter) Tokens() float64 {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.refill()
	return l.tokens
}

// SetLimit updates the rate limit dynamically
func (l *Limiter) SetLimit(calls int, window time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.maxTokens = float64(calls)
	l.refillRate = float64(calls) / window.Seconds()
	if l.tokens > l.maxTokens {
		l.tokens = l.maxTokens
	}
}

// Reset resets the limiter to full capacity
func (l *Limiter) Reset() {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.tokens = l.maxTokens
	l.lastRefill = time.Now()
}

// refill adds tokens based on elapsed time (must be called with lock held)
func (l *Limiter) refill() {
	now := time.Now()
	elapsed := now.Sub(l.lastRefill).Seconds()
	l.lastRefill = now

	l.tokens += elapsed * l.refillRate
	if l.tokens > l.maxTokens {
		l.tokens = l.maxTokens
	}
}

// timeUntilAvailable returns the time until at least one token is available
func (l *Limiter) timeUntilAvailable() time.Duration {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.refill()

	if l.tokens >= 1 {
		return 0
	}

	deficit := 1 - l.tokens
	return time.Duration(deficit/l.refillRate*1000) * time.Millisecond
}

// MultiLimiter manages multiple rate limiters for different platforms/endpoints
type MultiLimiter struct {
	mu       sync.RWMutex
	limiters map[string]*Limiter
}

// NewMultiLimiter creates a new multi-limiter
func NewMultiLimiter() *MultiLimiter {
	return &MultiLimiter{
		limiters: make(map[string]*Limiter),
	}
}

// GetOrCreate gets an existing limiter or creates a new one
func (ml *MultiLimiter) GetOrCreate(key string, calls int, window time.Duration) *Limiter {
	ml.mu.RLock()
	if limiter, ok := ml.limiters[key]; ok {
		ml.mu.RUnlock()
		return limiter
	}
	ml.mu.RUnlock()

	ml.mu.Lock()
	defer ml.mu.Unlock()

	// Double-check after acquiring write lock
	if limiter, ok := ml.limiters[key]; ok {
		return limiter
	}

	limiter := NewLimiter(calls, window)
	ml.limiters[key] = limiter
	return limiter
}

// Get gets a limiter by key
func (ml *MultiLimiter) Get(key string) (*Limiter, bool) {
	ml.mu.RLock()
	defer ml.mu.RUnlock()

	limiter, ok := ml.limiters[key]
	return limiter, ok
}

// Remove removes a limiter
func (ml *MultiLimiter) Remove(key string) {
	ml.mu.Lock()
	defer ml.mu.Unlock()

	delete(ml.limiters, key)
}

// Wait waits for a specific limiter
func (ml *MultiLimiter) Wait(ctx context.Context, key string) error {
	limiter, ok := ml.Get(key)
	if !ok {
		return nil // No limiter configured, allow
	}
	return limiter.Wait(ctx)
}

// Allow checks if a request is allowed for a specific limiter
func (ml *MultiLimiter) Allow(key string) bool {
	limiter, ok := ml.Get(key)
	if !ok {
		return true // No limiter configured, allow
	}
	return limiter.Allow()
}

// SlidingWindowLimiter implements a sliding window rate limiter
// This is more accurate than token bucket for strict rate limiting
type SlidingWindowLimiter struct {
	mu       sync.Mutex
	requests []time.Time
	maxCalls int
	window   time.Duration
}

// NewSlidingWindowLimiter creates a new sliding window rate limiter
func NewSlidingWindowLimiter(maxCalls int, window time.Duration) *SlidingWindowLimiter {
	return &SlidingWindowLimiter{
		requests: make([]time.Time, 0, maxCalls),
		maxCalls: maxCalls,
		window:   window,
	}
}

// Allow checks if a request is allowed
func (l *SlidingWindowLimiter) Allow() bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	threshold := now.Add(-l.window)

	// Remove expired requests
	newRequests := l.requests[:0]
	for _, t := range l.requests {
		if t.After(threshold) {
			newRequests = append(newRequests, t)
		}
	}
	l.requests = newRequests

	// Check if we can allow this request
	if len(l.requests) < l.maxCalls {
		l.requests = append(l.requests, now)
		return true
	}

	return false
}

// Wait waits until a request is allowed
func (l *SlidingWindowLimiter) Wait(ctx context.Context) error {
	for {
		if l.Allow() {
			return nil
		}

		waitTime := l.timeUntilAvailable()
		if waitTime == 0 {
			waitTime = 10 * time.Millisecond // Minimum wait
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(waitTime):
			continue
		}
	}
}

// timeUntilAvailable returns the time until the next request can be made
func (l *SlidingWindowLimiter) timeUntilAvailable() time.Duration {
	l.mu.Lock()
	defer l.mu.Unlock()

	if len(l.requests) < l.maxCalls {
		return 0
	}

	// Find the oldest request that will expire
	oldest := l.requests[0]
	expiresAt := oldest.Add(l.window)
	waitTime := time.Until(expiresAt)

	if waitTime < 0 {
		return 0
	}
	return waitTime
}

// Remaining returns the number of remaining requests allowed
func (l *SlidingWindowLimiter) Remaining() int {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	threshold := now.Add(-l.window)

	count := 0
	for _, t := range l.requests {
		if t.After(threshold) {
			count++
		}
	}

	remaining := l.maxCalls - count
	if remaining < 0 {
		return 0
	}
	return remaining
}
