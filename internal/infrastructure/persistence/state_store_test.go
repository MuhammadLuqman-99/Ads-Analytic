package persistence

import (
	"context"
	"testing"
	"time"

	"github.com/ads-aggregator/ads-aggregator/internal/domain/entity"
	"github.com/google/uuid"
)

func TestInMemoryStateStore_SaveAndGet(t *testing.T) {
	store := NewInMemoryStateStore(15 * time.Minute)
	ctx := context.Background()

	state := &entity.OAuthState{
		State:          "test_state_123",
		OrganizationID: uuid.New(),
		UserID:         uuid.New(),
		Platform:       entity.PlatformMeta,
		RedirectURL:    "/settings/connections",
		CreatedAt:      time.Now(),
		ExpiresAt:      time.Now().Add(10 * time.Minute),
	}

	// Save
	if err := store.Save(ctx, state); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// Get
	retrieved, err := store.Get(ctx, state.State)
	if err != nil {
		t.Fatalf("Get() failed: %v", err)
	}

	if retrieved.State != state.State {
		t.Errorf("State mismatch: got %q, want %q", retrieved.State, state.State)
	}

	if retrieved.Platform != state.Platform {
		t.Errorf("Platform mismatch: got %q, want %q", retrieved.Platform, state.Platform)
	}

	if retrieved.OrganizationID != state.OrganizationID {
		t.Errorf("OrganizationID mismatch")
	}

	if retrieved.UserID != state.UserID {
		t.Errorf("UserID mismatch")
	}

	if retrieved.RedirectURL != state.RedirectURL {
		t.Errorf("RedirectURL mismatch: got %q, want %q", retrieved.RedirectURL, state.RedirectURL)
	}
}

func TestInMemoryStateStore_Delete(t *testing.T) {
	store := NewInMemoryStateStore(15 * time.Minute)
	ctx := context.Background()

	state := &entity.OAuthState{
		State:          "test_state_to_delete",
		OrganizationID: uuid.New(),
		UserID:         uuid.New(),
		Platform:       entity.PlatformMeta,
		CreatedAt:      time.Now(),
		ExpiresAt:      time.Now().Add(10 * time.Minute),
	}

	// Save
	store.Save(ctx, state)

	// Verify exists
	if _, err := store.Get(ctx, state.State); err != nil {
		t.Fatalf("expected state to exist before delete")
	}

	// Delete
	if err := store.Delete(ctx, state.State); err != nil {
		t.Fatalf("Delete() failed: %v", err)
	}

	// Verify deleted
	_, err := store.Get(ctx, state.State)
	if err == nil {
		t.Error("expected error after deletion")
	}
}

func TestInMemoryStateStore_GetNonExistent(t *testing.T) {
	store := NewInMemoryStateStore(15 * time.Minute)
	ctx := context.Background()

	_, err := store.Get(ctx, "non_existent_state")
	if err == nil {
		t.Error("expected error for non-existent state")
	}
}

func TestInMemoryStateStore_SaveNilState(t *testing.T) {
	store := NewInMemoryStateStore(15 * time.Minute)
	ctx := context.Background()

	err := store.Save(ctx, nil)
	if err == nil {
		t.Error("expected error for nil state")
	}
}

func TestInMemoryStateStore_SaveEmptyStateID(t *testing.T) {
	store := NewInMemoryStateStore(15 * time.Minute)
	ctx := context.Background()

	state := &entity.OAuthState{
		State:    "",
		Platform: entity.PlatformMeta,
	}

	err := store.Save(ctx, state)
	if err == nil {
		t.Error("expected error for empty state ID")
	}
}

func TestInMemoryStateStore_GetEmptyStateID(t *testing.T) {
	store := NewInMemoryStateStore(15 * time.Minute)
	ctx := context.Background()

	_, err := store.Get(ctx, "")
	if err == nil {
		t.Error("expected error for empty state ID")
	}
}

func TestInMemoryStateStore_Expiry(t *testing.T) {
	// Create store with very short TTL
	store := NewInMemoryStateStore(50 * time.Millisecond)
	ctx := context.Background()

	state := &entity.OAuthState{
		State:          "test_expiring_state",
		OrganizationID: uuid.New(),
		UserID:         uuid.New(),
		Platform:       entity.PlatformMeta,
		CreatedAt:      time.Now(),
		ExpiresAt:      time.Now().Add(10 * time.Minute),
	}

	// Save
	store.Save(ctx, state)

	// Should exist immediately
	if _, err := store.Get(ctx, state.State); err != nil {
		t.Fatalf("expected state to exist immediately: %v", err)
	}

	// Wait for expiry
	time.Sleep(100 * time.Millisecond)

	// Should be expired
	_, err := store.Get(ctx, state.State)
	if err == nil {
		t.Error("expected state to be expired")
	}
}

func TestInMemoryStateStore_Count(t *testing.T) {
	store := NewInMemoryStateStore(15 * time.Minute)
	ctx := context.Background()

	// Initially empty
	if count := store.Count(); count != 0 {
		t.Errorf("expected 0, got %d", count)
	}

	// Add states
	for i := 0; i < 5; i++ {
		state := &entity.OAuthState{
			State:    uuid.NewString(),
			Platform: entity.PlatformMeta,
		}
		store.Save(ctx, state)
	}

	if count := store.Count(); count != 5 {
		t.Errorf("expected 5, got %d", count)
	}
}

func TestInMemoryStateStore_Clear(t *testing.T) {
	store := NewInMemoryStateStore(15 * time.Minute)
	ctx := context.Background()

	// Add states
	for i := 0; i < 5; i++ {
		state := &entity.OAuthState{
			State:    uuid.NewString(),
			Platform: entity.PlatformMeta,
		}
		store.Save(ctx, state)
	}

	// Clear
	store.Clear()

	if count := store.Count(); count != 0 {
		t.Errorf("expected 0 after clear, got %d", count)
	}
}

func TestInMemoryStateStore_ConcurrentAccess(t *testing.T) {
	store := NewInMemoryStateStore(15 * time.Minute)
	ctx := context.Background()

	done := make(chan bool)

	// Writer goroutine
	go func() {
		for i := 0; i < 100; i++ {
			state := &entity.OAuthState{
				State:    uuid.NewString(),
				Platform: entity.PlatformMeta,
			}
			store.Save(ctx, state)
		}
		done <- true
	}()

	// Reader goroutine
	go func() {
		for i := 0; i < 100; i++ {
			store.Count()
		}
		done <- true
	}()

	// Wait for both
	<-done
	<-done
}

func TestInMemoryStateStore_Overwrite(t *testing.T) {
	store := NewInMemoryStateStore(15 * time.Minute)
	ctx := context.Background()

	stateID := "same_state_id"

	// Save first state
	state1 := &entity.OAuthState{
		State:       stateID,
		Platform:    entity.PlatformMeta,
		RedirectURL: "/first",
	}
	store.Save(ctx, state1)

	// Save second state with same ID
	state2 := &entity.OAuthState{
		State:       stateID,
		Platform:    entity.PlatformTikTok,
		RedirectURL: "/second",
	}
	store.Save(ctx, state2)

	// Should get the latest
	retrieved, err := store.Get(ctx, stateID)
	if err != nil {
		t.Fatalf("Get() failed: %v", err)
	}

	if retrieved.Platform != entity.PlatformTikTok {
		t.Errorf("expected %q, got %q", entity.PlatformTikTok, retrieved.Platform)
	}

	if retrieved.RedirectURL != "/second" {
		t.Errorf("expected %q, got %q", "/second", retrieved.RedirectURL)
	}
}

func TestInMemoryStateStore_DefaultTTL(t *testing.T) {
	// Create with zero TTL (should default to 15 minutes)
	store := NewInMemoryStateStore(0)

	if store.ttl != 15*time.Minute {
		t.Errorf("expected default TTL of 15 minutes, got %v", store.ttl)
	}
}

func BenchmarkStateStore_Save(b *testing.B) {
	store := NewInMemoryStateStore(15 * time.Minute)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		state := &entity.OAuthState{
			State:    uuid.NewString(),
			Platform: entity.PlatformMeta,
		}
		store.Save(ctx, state)
	}
}

func BenchmarkStateStore_Get(b *testing.B) {
	store := NewInMemoryStateStore(15 * time.Minute)
	ctx := context.Background()

	stateID := "benchmark_state"
	state := &entity.OAuthState{
		State:    stateID,
		Platform: entity.PlatformMeta,
	}
	store.Save(ctx, state)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.Get(ctx, stateID)
	}
}
