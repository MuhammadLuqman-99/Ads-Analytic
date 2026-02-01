package metrics

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics holds all application metrics
type Metrics struct {
	// HTTP metrics
	HTTPRequestsTotal    *prometheus.CounterVec
	HTTPRequestDuration  *prometheus.HistogramVec
	HTTPRequestsInFlight prometheus.Gauge
	HTTPResponseSize     *prometheus.HistogramVec

	// Business metrics
	ActiveUsers      prometheus.Gauge
	ActiveOrgs       prometheus.Gauge
	ConnectedAccounts *prometheus.GaugeVec

	// Sync job metrics
	SyncJobsTotal     *prometheus.CounterVec
	SyncJobDuration   *prometheus.HistogramVec
	SyncJobsInFlight  *prometheus.GaugeVec
	RecordsSynced     *prometheus.CounterVec
	SyncErrors        *prometheus.CounterVec

	// Platform API metrics
	PlatformAPICallsTotal    *prometheus.CounterVec
	PlatformAPICallDuration  *prometheus.HistogramVec
	PlatformAPIErrors        *prometheus.CounterVec
	PlatformRateLimitHits    *prometheus.CounterVec

	// Database metrics
	DBQueryDuration  *prometheus.HistogramVec
	DBConnectionsOpen prometheus.Gauge
	DBConnectionsIdle prometheus.Gauge

	// Redis metrics
	RedisOperationDuration *prometheus.HistogramVec
	RedisErrors           *prometheus.CounterVec

	// Billing metrics
	SubscriptionsByPlan *prometheus.GaugeVec
	PaymentsTotal       *prometheus.CounterVec
	PaymentAmount       *prometheus.HistogramVec

	// Error metrics
	ErrorsTotal *prometheus.CounterVec
}

var (
	defaultMetrics *Metrics
	namespace      = "ads_analytics"
)

// Init initializes the metrics
func Init() *Metrics {
	defaultMetrics = &Metrics{
		// HTTP metrics
		HTTPRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "http_requests_total",
				Help:      "Total number of HTTP requests",
			},
			[]string{"method", "path", "status"},
		),
		HTTPRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "http_request_duration_seconds",
				Help:      "HTTP request duration in seconds",
				Buckets:   []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
			},
			[]string{"method", "path", "status"},
		),
		HTTPRequestsInFlight: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "http_requests_in_flight",
				Help:      "Current number of HTTP requests being processed",
			},
		),
		HTTPResponseSize: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "http_response_size_bytes",
				Help:      "HTTP response size in bytes",
				Buckets:   prometheus.ExponentialBuckets(100, 10, 8),
			},
			[]string{"method", "path"},
		),

		// Business metrics
		ActiveUsers: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "active_users",
				Help:      "Number of active users in the last 24 hours",
			},
		),
		ActiveOrgs: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "active_organizations",
				Help:      "Number of active organizations",
			},
		),
		ConnectedAccounts: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "connected_accounts",
				Help:      "Number of connected ad platform accounts",
			},
			[]string{"platform", "status"},
		),

		// Sync job metrics
		SyncJobsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "sync_jobs_total",
				Help:      "Total number of sync jobs executed",
			},
			[]string{"platform", "type", "status"},
		),
		SyncJobDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "sync_job_duration_seconds",
				Help:      "Sync job duration in seconds",
				Buckets:   []float64{1, 5, 10, 30, 60, 120, 300, 600},
			},
			[]string{"platform", "type"},
		),
		SyncJobsInFlight: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "sync_jobs_in_flight",
				Help:      "Current number of sync jobs running",
			},
			[]string{"platform"},
		),
		RecordsSynced: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "records_synced_total",
				Help:      "Total number of records synced",
			},
			[]string{"platform", "type"},
		),
		SyncErrors: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "sync_errors_total",
				Help:      "Total number of sync errors",
			},
			[]string{"platform", "type", "error_type"},
		),

		// Platform API metrics
		PlatformAPICallsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "platform_api_calls_total",
				Help:      "Total number of platform API calls",
			},
			[]string{"platform", "endpoint", "status"},
		),
		PlatformAPICallDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "platform_api_call_duration_seconds",
				Help:      "Platform API call duration in seconds",
				Buckets:   []float64{.1, .25, .5, 1, 2.5, 5, 10, 30},
			},
			[]string{"platform", "endpoint"},
		),
		PlatformAPIErrors: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "platform_api_errors_total",
				Help:      "Total number of platform API errors",
			},
			[]string{"platform", "endpoint", "error_code"},
		),
		PlatformRateLimitHits: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "platform_rate_limit_hits_total",
				Help:      "Total number of rate limit hits",
			},
			[]string{"platform"},
		),

		// Database metrics
		DBQueryDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "db_query_duration_seconds",
				Help:      "Database query duration in seconds",
				Buckets:   []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
			},
			[]string{"operation", "table"},
		),
		DBConnectionsOpen: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "db_connections_open",
				Help:      "Number of open database connections",
			},
		),
		DBConnectionsIdle: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "db_connections_idle",
				Help:      "Number of idle database connections",
			},
		),

		// Redis metrics
		RedisOperationDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "redis_operation_duration_seconds",
				Help:      "Redis operation duration in seconds",
				Buckets:   []float64{.0001, .0005, .001, .005, .01, .025, .05, .1},
			},
			[]string{"operation"},
		),
		RedisErrors: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "redis_errors_total",
				Help:      "Total number of Redis errors",
			},
			[]string{"operation"},
		),

		// Billing metrics
		SubscriptionsByPlan: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "subscriptions_by_plan",
				Help:      "Number of subscriptions by plan",
			},
			[]string{"plan", "status"},
		),
		PaymentsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "payments_total",
				Help:      "Total number of payments",
			},
			[]string{"status", "plan"},
		),
		PaymentAmount: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "payment_amount_myr",
				Help:      "Payment amount in MYR",
				Buckets:   []float64{10, 50, 100, 200, 500, 1000, 2000, 5000},
			},
			[]string{"plan"},
		),

		// Error metrics
		ErrorsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "errors_total",
				Help:      "Total number of errors",
			},
			[]string{"type", "severity", "component"},
		),
	}

	return defaultMetrics
}

// Default returns the default metrics instance
func Default() *Metrics {
	if defaultMetrics == nil {
		Init()
	}
	return defaultMetrics
}

// GinMiddleware returns a Gin middleware for metrics collection
func GinMiddleware() gin.HandlerFunc {
	m := Default()

	return func(c *gin.Context) {
		// Skip metrics endpoint
		if c.Request.URL.Path == "/metrics" {
			c.Next()
			return
		}

		start := time.Now()
		m.HTTPRequestsInFlight.Inc()

		// Process request
		c.Next()

		m.HTTPRequestsInFlight.Dec()
		duration := time.Since(start).Seconds()

		status := strconv.Itoa(c.Writer.Status())
		path := c.FullPath()
		if path == "" {
			path = "not_found"
		}

		// Record metrics
		m.HTTPRequestsTotal.WithLabelValues(c.Request.Method, path, status).Inc()
		m.HTTPRequestDuration.WithLabelValues(c.Request.Method, path, status).Observe(duration)
		m.HTTPResponseSize.WithLabelValues(c.Request.Method, path).Observe(float64(c.Writer.Size()))
	}
}

// Handler returns the Prometheus HTTP handler
func Handler() gin.HandlerFunc {
	h := promhttp.Handler()
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

// ============================================================================
// Helper functions for recording metrics
// ============================================================================

// RecordSyncJob records a sync job execution
func (m *Metrics) RecordSyncJob(platform, jobType, status string, duration time.Duration, recordCount int) {
	m.SyncJobsTotal.WithLabelValues(platform, jobType, status).Inc()
	m.SyncJobDuration.WithLabelValues(platform, jobType).Observe(duration.Seconds())
	if recordCount > 0 {
		m.RecordsSynced.WithLabelValues(platform, jobType).Add(float64(recordCount))
	}
}

// RecordSyncError records a sync error
func (m *Metrics) RecordSyncError(platform, jobType, errorType string) {
	m.SyncErrors.WithLabelValues(platform, jobType, errorType).Inc()
}

// RecordPlatformAPICall records a platform API call
func (m *Metrics) RecordPlatformAPICall(platform, endpoint, status string, duration time.Duration) {
	m.PlatformAPICallsTotal.WithLabelValues(platform, endpoint, status).Inc()
	m.PlatformAPICallDuration.WithLabelValues(platform, endpoint).Observe(duration.Seconds())
}

// RecordPlatformAPIError records a platform API error
func (m *Metrics) RecordPlatformAPIError(platform, endpoint, errorCode string) {
	m.PlatformAPIErrors.WithLabelValues(platform, endpoint, errorCode).Inc()
}

// RecordRateLimitHit records a rate limit hit
func (m *Metrics) RecordRateLimitHit(platform string) {
	m.PlatformRateLimitHits.WithLabelValues(platform).Inc()
}

// RecordDBQuery records a database query
func (m *Metrics) RecordDBQuery(operation, table string, duration time.Duration) {
	m.DBQueryDuration.WithLabelValues(operation, table).Observe(duration.Seconds())
}

// RecordRedisOperation records a Redis operation
func (m *Metrics) RecordRedisOperation(operation string, duration time.Duration) {
	m.RedisOperationDuration.WithLabelValues(operation).Observe(duration.Seconds())
}

// RecordRedisError records a Redis error
func (m *Metrics) RecordRedisError(operation string) {
	m.RedisErrors.WithLabelValues(operation).Inc()
}

// RecordPayment records a payment
func (m *Metrics) RecordPayment(status, plan string, amount float64) {
	m.PaymentsTotal.WithLabelValues(status, plan).Inc()
	if status == "succeeded" {
		m.PaymentAmount.WithLabelValues(plan).Observe(amount)
	}
}

// RecordError records an error
func (m *Metrics) RecordError(errorType, severity, component string) {
	m.ErrorsTotal.WithLabelValues(errorType, severity, component).Inc()
}

// UpdateActiveUsers updates the active users gauge
func (m *Metrics) UpdateActiveUsers(count float64) {
	m.ActiveUsers.Set(count)
}

// UpdateActiveOrgs updates the active organizations gauge
func (m *Metrics) UpdateActiveOrgs(count float64) {
	m.ActiveOrgs.Set(count)
}

// UpdateConnectedAccounts updates the connected accounts gauge
func (m *Metrics) UpdateConnectedAccounts(platform, status string, count float64) {
	m.ConnectedAccounts.WithLabelValues(platform, status).Set(count)
}

// UpdateSubscriptions updates the subscriptions gauge
func (m *Metrics) UpdateSubscriptions(plan, status string, count float64) {
	m.SubscriptionsByPlan.WithLabelValues(plan, status).Set(count)
}

// UpdateDBConnections updates database connection gauges
func (m *Metrics) UpdateDBConnections(open, idle int) {
	m.DBConnectionsOpen.Set(float64(open))
	m.DBConnectionsIdle.Set(float64(idle))
}

// StartSyncJob marks a sync job as started
func (m *Metrics) StartSyncJob(platform string) {
	m.SyncJobsInFlight.WithLabelValues(platform).Inc()
}

// EndSyncJob marks a sync job as ended
func (m *Metrics) EndSyncJob(platform string) {
	m.SyncJobsInFlight.WithLabelValues(platform).Dec()
}
