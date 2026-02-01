package persistence

import (
	"context"
	"sync"
	"time"

	"github.com/ads-aggregator/ads-aggregator/internal/domain/entity"
	"github.com/ads-aggregator/ads-aggregator/pkg/errors"
)

// InMemoryStateStore provides an in-memory implementation of StateStore for OAuth flows
// This is suitable for single-instance deployments. For distributed systems,
// consider using Redis or a database-backed implementation.
type InMemoryStateStore struct {
	mu     sync.RWMutex
	states map[string]*stateEntry
	ttl    time.Duration
}

type stateEntry struct {
	state     *entity.OAuthState
	createdAt time.Time
}

// NewInMemoryStateStore creates a new in-memory state store
// Default TTL is 15 minutes if not specified
func NewInMemoryStateStore(ttl time.Duration) *InMemoryStateStore {
	if ttl == 0 {
		ttl = 15 * time.Minute
	}

	store := &InMemoryStateStore{
		states: make(map[string]*stateEntry),
		ttl:    ttl,
	}

	// Start cleanup goroutine
	go store.cleanupLoop()

	return store
}

// Save stores an OAuth state
func (s *InMemoryStateStore) Save(ctx context.Context, state *entity.OAuthState) error {
	if state == nil {
		return errors.ErrBadRequest("state cannot be nil")
	}

	if state.State == "" {
		return errors.ErrBadRequest("state ID cannot be empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.states[state.State] = &stateEntry{
		state:     state,
		createdAt: time.Now(),
	}

	return nil
}

// Get retrieves an OAuth state by its ID
func (s *InMemoryStateStore) Get(ctx context.Context, stateID string) (*entity.OAuthState, error) {
	if stateID == "" {
		return nil, errors.ErrBadRequest("state ID cannot be empty")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	entry, exists := s.states[stateID]
	if !exists {
		return nil, errors.ErrNotFound("OAuth state")
	}

	// Check if expired based on internal TTL
	if time.Since(entry.createdAt) > s.ttl {
		return nil, errors.ErrNotFound("OAuth state (expired)")
	}

	return entry.state, nil
}

// Delete removes an OAuth state
func (s *InMemoryStateStore) Delete(ctx context.Context, stateID string) error {
	if stateID == "" {
		return errors.ErrBadRequest("state ID cannot be empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.states, stateID)
	return nil
}

// cleanupLoop periodically removes expired states
func (s *InMemoryStateStore) cleanupLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.cleanup()
	}
}

// cleanup removes expired entries
func (s *InMemoryStateStore) cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for id, entry := range s.states {
		if now.Sub(entry.createdAt) > s.ttl {
			delete(s.states, id)
		}
	}
}

// Count returns the number of stored states (for testing)
func (s *InMemoryStateStore) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.states)
}

// Clear removes all states (for testing)
func (s *InMemoryStateStore) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.states = make(map[string]*stateEntry)
}
