package events

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/google/uuid"
)

// EventType represents the type of event
type EventType string

const (
	EventSyncStarted   EventType = "sync:started"
	EventSyncProgress  EventType = "sync:progress"
	EventSyncCompleted EventType = "sync:completed"
	EventSyncError     EventType = "sync:error"
	EventDataUpdated   EventType = "data:updated"
)

// Event represents a server-sent event
type Event struct {
	ID        string      `json:"id"`
	Type      EventType   `json:"type"`
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data"`
}

// SyncStartedData contains data for sync:started event
type SyncStartedData struct {
	Platform  string `json:"platform"`
	AccountID string `json:"account_id"`
}

// SyncProgressData contains data for sync:progress event
type SyncProgressData struct {
	Platform        string `json:"platform"`
	AccountID       string `json:"account_id"`
	ProgressPercent int    `json:"progress_percent"`
	Message         string `json:"message,omitempty"`
}

// SyncCompletedData contains data for sync:completed event
type SyncCompletedData struct {
	Platform      string        `json:"platform"`
	AccountID     string        `json:"account_id"`
	RecordsSynced int           `json:"records_synced"`
	Duration      time.Duration `json:"duration"`
	DurationStr   string        `json:"duration_str"`
}

// SyncErrorData contains data for sync:error event
type SyncErrorData struct {
	Platform     string `json:"platform"`
	AccountID    string `json:"account_id"`
	ErrorMessage string `json:"error_message"`
	Retryable    bool   `json:"retryable"`
}

// DataUpdatedData contains data for data:updated event
type DataUpdatedData struct {
	Affected []string `json:"affected"` // e.g., ["dashboard", "campaigns"]
}

// Client represents a connected SSE client
type Client struct {
	ID       string
	OrgID    uuid.UUID
	UserID   uuid.UUID
	Channel  chan []byte
	Created  time.Time
}

// Broadcaster manages SSE clients and event broadcasting
type Broadcaster struct {
	mu      sync.RWMutex
	clients map[string]*Client
}

// NewBroadcaster creates a new event broadcaster
func NewBroadcaster() *Broadcaster {
	return &Broadcaster{
		clients: make(map[string]*Client),
	}
}

// Register adds a new client to the broadcaster
func (b *Broadcaster) Register(client *Client) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.clients[client.ID] = client
}

// Unregister removes a client from the broadcaster
func (b *Broadcaster) Unregister(clientID string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if client, ok := b.clients[clientID]; ok {
		close(client.Channel)
		delete(b.clients, clientID)
	}
}

// BroadcastToOrg sends an event to all clients in an organization
func (b *Broadcaster) BroadcastToOrg(orgID uuid.UUID, event Event) {
	data, err := json.Marshal(event)
	if err != nil {
		return
	}

	// Format as SSE
	sseData := formatSSE(string(event.Type), data)

	b.mu.RLock()
	defer b.mu.RUnlock()

	for _, client := range b.clients {
		if client.OrgID == orgID {
			select {
			case client.Channel <- sseData:
			default:
				// Client channel full, skip
			}
		}
	}
}

// BroadcastToUser sends an event to a specific user
func (b *Broadcaster) BroadcastToUser(userID uuid.UUID, event Event) {
	data, err := json.Marshal(event)
	if err != nil {
		return
	}

	sseData := formatSSE(string(event.Type), data)

	b.mu.RLock()
	defer b.mu.RUnlock()

	for _, client := range b.clients {
		if client.UserID == userID {
			select {
			case client.Channel <- sseData:
			default:
			}
		}
	}
}

// ClientCount returns the number of connected clients
func (b *Broadcaster) ClientCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.clients)
}

// ClientCountForOrg returns the number of connected clients for an organization
func (b *Broadcaster) ClientCountForOrg(orgID uuid.UUID) int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	count := 0
	for _, client := range b.clients {
		if client.OrgID == orgID {
			count++
		}
	}
	return count
}

// formatSSE formats data as Server-Sent Events format
func formatSSE(eventType string, data []byte) []byte {
	return []byte("event: " + eventType + "\ndata: " + string(data) + "\n\n")
}

// NewEvent creates a new event with auto-generated ID and timestamp
func NewEvent(eventType EventType, data interface{}) Event {
	return Event{
		ID:        uuid.New().String(),
		Type:      eventType,
		Timestamp: time.Now().UTC(),
		Data:      data,
	}
}

// Helper functions to create specific events

// NewSyncStartedEvent creates a sync:started event
func NewSyncStartedEvent(platform, accountID string) Event {
	return NewEvent(EventSyncStarted, SyncStartedData{
		Platform:  platform,
		AccountID: accountID,
	})
}

// NewSyncProgressEvent creates a sync:progress event
func NewSyncProgressEvent(platform, accountID string, progress int, message string) Event {
	return NewEvent(EventSyncProgress, SyncProgressData{
		Platform:        platform,
		AccountID:       accountID,
		ProgressPercent: progress,
		Message:         message,
	})
}

// NewSyncCompletedEvent creates a sync:completed event
func NewSyncCompletedEvent(platform, accountID string, recordsSynced int, duration time.Duration) Event {
	return NewEvent(EventSyncCompleted, SyncCompletedData{
		Platform:      platform,
		AccountID:     accountID,
		RecordsSynced: recordsSynced,
		Duration:      duration,
		DurationStr:   duration.String(),
	})
}

// NewSyncErrorEvent creates a sync:error event
func NewSyncErrorEvent(platform, accountID, errorMessage string, retryable bool) Event {
	return NewEvent(EventSyncError, SyncErrorData{
		Platform:     platform,
		AccountID:    accountID,
		ErrorMessage: errorMessage,
		Retryable:    retryable,
	})
}

// NewDataUpdatedEvent creates a data:updated event
func NewDataUpdatedEvent(affected []string) Event {
	return NewEvent(EventDataUpdated, DataUpdatedData{
		Affected: affected,
	})
}
