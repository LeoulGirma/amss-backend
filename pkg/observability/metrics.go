package observability

import (
	"strconv"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc/codes"
)

var registerOnce sync.Once

var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests.",
		},
		[]string{"route", "method", "status"},
	)
	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request latency in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"route", "method", "status"},
	)
	grpcRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "grpc_requests_total",
			Help: "Total number of gRPC requests.",
		},
		[]string{"method", "code"},
	)
	grpcRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "grpc_request_duration_seconds",
			Help:    "gRPC request latency in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "code"},
	)
	jobRunsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "job_runs_total",
			Help: "Total number of job runs.",
		},
		[]string{"job"},
	)
	jobFailuresTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "job_failures_total",
			Help: "Total number of job failures.",
		},
		[]string{"job"},
	)
	outboxProcessedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "outbox_events_processed_total",
			Help: "Total number of outbox events processed.",
		},
		[]string{"event_type"},
	)
	outboxFailedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "outbox_events_failed_total",
			Help: "Total number of outbox event failures.",
		},
		[]string{"event_type"},
	)
	webhookDeliveriesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "webhook_deliveries_total",
			Help: "Total number of webhook deliveries by status.",
		},
		[]string{"status"},
	)
	rateLimitRejectionsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rate_limit_rejections_total",
			Help: "Total number of rate limit rejections.",
		},
		[]string{"scope"},
	)
)

func RegisterMetrics() {
	registerOnce.Do(func() {
		prometheus.MustRegister(
			httpRequestsTotal,
			httpRequestDuration,
			grpcRequestsTotal,
			grpcRequestDuration,
			jobRunsTotal,
			jobFailuresTotal,
			outboxProcessedTotal,
			outboxFailedTotal,
			webhookDeliveriesTotal,
			rateLimitRejectionsTotal,
		)
	})
}

func ObserveHTTPRequest(route, method string, status int, duration time.Duration) {
	code := strconv.Itoa(status)
	httpRequestsTotal.WithLabelValues(route, method, code).Inc()
	httpRequestDuration.WithLabelValues(route, method, code).Observe(duration.Seconds())
}

func ObserveGRPC(method string, code codes.Code, duration time.Duration) {
	label := code.String()
	grpcRequestsTotal.WithLabelValues(method, label).Inc()
	grpcRequestDuration.WithLabelValues(method, label).Observe(duration.Seconds())
}

func IncJobRun(job string) {
	jobRunsTotal.WithLabelValues(job).Inc()
}

func IncJobFailure(job string) {
	jobFailuresTotal.WithLabelValues(job).Inc()
}

func IncOutboxProcessed(eventType string) {
	outboxProcessedTotal.WithLabelValues(eventType).Inc()
}

func IncOutboxFailed(eventType string) {
	outboxFailedTotal.WithLabelValues(eventType).Inc()
}

func IncWebhookDelivery(status string) {
	webhookDeliveriesTotal.WithLabelValues(status).Inc()
}

func IncRateLimitRejection(scope string) {
	rateLimitRejectionsTotal.WithLabelValues(scope).Inc()
}
