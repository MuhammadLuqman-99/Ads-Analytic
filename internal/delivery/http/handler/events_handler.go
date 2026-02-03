package handler

import (
	"net/http"
	"time"

	"github.com/ads-aggregator/ads-aggregator/internal/delivery/http/middleware"
	"github.com/ads-aggregator/ads-aggregator/internal/infrastructure/events"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// EventsHandler handles SSE event streaming
type EventsHandler struct {
	broadcaster *events.Broadcaster
}

// NewEventsHandler creates a new events handler
func NewEventsHandler(broadcaster *events.Broadcaster) *EventsHandler {
	return &EventsHandler{broadcaster: broadcaster}
}

// Stream handles GET /api/v1/events/stream
// Establishes an SSE connection for real-time updates
func (h *EventsHandler) Stream(c *gin.Context) {
	// Get user context
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	orgID, ok := middleware.GetOrgID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Organization not found in context"})
		return
	}

	// Set SSE headers
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	// Note: CORS headers are handled by the CORS middleware
	// Do NOT set Access-Control-Allow-Origin: * as it conflicts with credentials
	c.Writer.Header().Set("X-Accel-Buffering", "no") // Disable nginx buffering

	// Create client
	client := &events.Client{
		ID:      uuid.New().String(),
		OrgID:   orgID,
		UserID:  userID,
		Channel: make(chan []byte, 100), // Buffered channel
		Created: time.Now(),
	}

	// Register client
	h.broadcaster.Register(client)
	defer h.broadcaster.Unregister(client.ID)

	// Send initial connection event
	connectedData := []byte("event: connected\ndata: {\"client_id\":\"" + client.ID + "\"}\n\n")
	c.Writer.Write(connectedData)
	c.Writer.Flush()

	// Create heartbeat ticker (every 30 seconds to keep connection alive)
	heartbeat := time.NewTicker(30 * time.Second)
	defer heartbeat.Stop()

	// Listen for events
	clientGone := c.Request.Context().Done()
	for {
		select {
		case <-clientGone:
			// Client disconnected
			return
		case data := <-client.Channel:
			// Send event to client
			c.Writer.Write(data)
			c.Writer.Flush()
		case <-heartbeat.C:
			// Send heartbeat to keep connection alive
			c.Writer.Write([]byte(": heartbeat\n\n"))
			c.Writer.Flush()
		}
	}
}

// GetStatus returns the current sync status for all connected accounts
// GET /api/v1/events/status
func (h *EventsHandler) GetStatus(c *gin.Context) {
	orgID, ok := middleware.GetOrgID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Organization not found in context"})
		return
	}

	// Return connected client count for this org (useful for debugging)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"connected_clients": h.broadcaster.ClientCountForOrg(orgID),
			"timestamp":         time.Now().UTC().Format(time.RFC3339),
		},
	})
}

// BroadcastSyncStarted broadcasts a sync started event
func (h *EventsHandler) BroadcastSyncStarted(orgID uuid.UUID, platform, accountID string) {
	event := events.NewSyncStartedEvent(platform, accountID)
	h.broadcaster.BroadcastToOrg(orgID, event)
}

// BroadcastSyncProgress broadcasts a sync progress event
func (h *EventsHandler) BroadcastSyncProgress(orgID uuid.UUID, platform, accountID string, progress int, message string) {
	event := events.NewSyncProgressEvent(platform, accountID, progress, message)
	h.broadcaster.BroadcastToOrg(orgID, event)
}

// BroadcastSyncCompleted broadcasts a sync completed event
func (h *EventsHandler) BroadcastSyncCompleted(orgID uuid.UUID, platform, accountID string, recordsSynced int, duration time.Duration) {
	event := events.NewSyncCompletedEvent(platform, accountID, recordsSynced, duration)
	h.broadcaster.BroadcastToOrg(orgID, event)
}

// BroadcastSyncError broadcasts a sync error event
func (h *EventsHandler) BroadcastSyncError(orgID uuid.UUID, platform, accountID, errorMessage string, retryable bool) {
	event := events.NewSyncErrorEvent(platform, accountID, errorMessage, retryable)
	h.broadcaster.BroadcastToOrg(orgID, event)
}

// BroadcastDataUpdated broadcasts a data updated event
func (h *EventsHandler) BroadcastDataUpdated(orgID uuid.UUID, affected []string) {
	event := events.NewDataUpdatedEvent(affected)
	h.broadcaster.BroadcastToOrg(orgID, event)
}

// GetBroadcaster returns the broadcaster instance (for use by other services)
func (h *EventsHandler) GetBroadcaster() *events.Broadcaster {
	return h.broadcaster
}
