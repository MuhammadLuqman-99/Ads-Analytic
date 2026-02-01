package handler

import (
	"context"
	"database/sql"
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// HealthHandler handles health check endpoints
type HealthHandler struct {
	db          *sql.DB
	redisClient *redis.Client
	startTime   time.Time
	version     string
	gitCommit   string
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(db *sql.DB, redisClient *redis.Client, version, gitCommit string) *HealthHandler {
	return &HealthHandler{
		db:          db,
		redisClient: redisClient,
		startTime:   time.Now(),
		version:     version,
		gitCommit:   gitCommit,
	}
}

// HealthStatus represents the health check response
type HealthStatus struct {
	Status    string            `json:"status"`
	Version   string            `json:"version,omitempty"`
	GitCommit string            `json:"git_commit,omitempty"`
	Uptime    string            `json:"uptime,omitempty"`
	Timestamp string            `json:"timestamp"`
	Checks    map[string]Check  `json:"checks,omitempty"`
}

// Check represents an individual health check
type Check struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Latency string `json:"latency,omitempty"`
}

// HandleHealth returns basic health status (for load balancers)
func (h *HealthHandler) HandleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// HandleLiveness returns liveness probe status (is the service running?)
func (h *HealthHandler) HandleLiveness(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "alive",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// HandleReadiness returns readiness probe status (can the service handle requests?)
func (h *HealthHandler) HandleReadiness(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	checks := make(map[string]Check)
	allHealthy := true

	// Check database
	if h.db != nil {
		dbCheck := h.checkDatabase(ctx)
		checks["database"] = dbCheck
		if dbCheck.Status != "healthy" {
			allHealthy = false
		}
	}

	// Check Redis
	if h.redisClient != nil {
		redisCheck := h.checkRedis(ctx)
		checks["redis"] = redisCheck
		if redisCheck.Status != "healthy" {
			allHealthy = false
		}
	}

	status := "ready"
	httpStatus := http.StatusOK
	if !allHealthy {
		status = "not_ready"
		httpStatus = http.StatusServiceUnavailable
	}

	c.JSON(httpStatus, HealthStatus{
		Status:    status,
		Version:   h.version,
		GitCommit: h.gitCommit,
		Uptime:    time.Since(h.startTime).String(),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Checks:    checks,
	})
}

// HandleDetailed returns detailed health status with all dependencies
func (h *HealthHandler) HandleDetailed(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	checks := make(map[string]Check)
	allHealthy := true

	// Check database
	if h.db != nil {
		dbCheck := h.checkDatabase(ctx)
		checks["database"] = dbCheck
		if dbCheck.Status != "healthy" {
			allHealthy = false
		}
	}

	// Check Redis
	if h.redisClient != nil {
		redisCheck := h.checkRedis(ctx)
		checks["redis"] = redisCheck
		if redisCheck.Status != "healthy" {
			allHealthy = false
		}
	}

	// Add system info
	checks["memory"] = h.checkMemory()
	checks["goroutines"] = Check{
		Status:  "healthy",
		Message: formatInt(runtime.NumGoroutine()) + " goroutines",
	}

	status := "healthy"
	httpStatus := http.StatusOK
	if !allHealthy {
		status = "unhealthy"
		httpStatus = http.StatusServiceUnavailable
	}

	c.JSON(httpStatus, HealthStatus{
		Status:    status,
		Version:   h.version,
		GitCommit: h.gitCommit,
		Uptime:    time.Since(h.startTime).String(),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Checks:    checks,
	})
}

// checkDatabase checks database connectivity
func (h *HealthHandler) checkDatabase(ctx context.Context) Check {
	start := time.Now()

	if err := h.db.PingContext(ctx); err != nil {
		return Check{
			Status:  "unhealthy",
			Message: err.Error(),
			Latency: time.Since(start).String(),
		}
	}

	return Check{
		Status:  "healthy",
		Message: "connected",
		Latency: time.Since(start).String(),
	}
}

// checkRedis checks Redis connectivity
func (h *HealthHandler) checkRedis(ctx context.Context) Check {
	start := time.Now()

	if err := h.redisClient.Ping(ctx).Err(); err != nil {
		return Check{
			Status:  "unhealthy",
			Message: err.Error(),
			Latency: time.Since(start).String(),
		}
	}

	return Check{
		Status:  "healthy",
		Message: "connected",
		Latency: time.Since(start).String(),
	}
}

// checkMemory returns memory usage status
func (h *HealthHandler) checkMemory() Check {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	allocMB := float64(m.Alloc) / 1024 / 1024
	sysMB := float64(m.Sys) / 1024 / 1024

	status := "healthy"
	if allocMB > 500 { // Warning if using more than 500MB
		status = "warning"
	}

	return Check{
		Status:  status,
		Message: formatFloat(allocMB, 2) + "MB / " + formatFloat(sysMB, 2) + "MB",
	}
}

// formatInt formats an integer as a string
func formatInt(n int) string {
	return string(rune('0'+n%10)) + formatIntRec(n/10)
}

func formatIntRec(n int) string {
	if n == 0 {
		return ""
	}
	return formatIntRec(n/10) + string(rune('0'+n%10))
}

// formatFloat formats a float with specified precision
func formatFloat(f float64, precision int) string {
	format := "%." + string(rune('0'+precision)) + "f"
	return sprintf(format, f)
}

// sprintf is a simple implementation
func sprintf(format string, f float64) string {
	// Simple implementation for our use case
	intPart := int(f)
	decPart := int((f - float64(intPart)) * 100)
	if decPart < 0 {
		decPart = -decPart
	}
	result := ""
	if intPart == 0 {
		result = "0"
	} else {
		for intPart > 0 {
			result = string(rune('0'+intPart%10)) + result
			intPart /= 10
		}
	}
	result += "."
	if decPart < 10 {
		result += "0"
	}
	decStr := ""
	for decPart > 0 {
		decStr = string(rune('0'+decPart%10)) + decStr
		decPart /= 10
	}
	if decStr == "" {
		decStr = "00"
	}
	return result + decStr
}
