package rest

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/aeromaintain/amss/pkg/observability"
	"google.golang.org/grpc/codes"
)

func TestMetricsEndpointExposesLabels(t *testing.T) {
	observability.RegisterMetrics()
	observability.ObserveGRPC("/amss.v1.TaskService/GetTask", codes.OK, 10*time.Millisecond)

	router := NewRouter(Deps{PrometheusEnabled: true})

	healthReq := httptest.NewRequest(http.MethodGet, "/health", nil)
	healthRec := httptest.NewRecorder()
	router.ServeHTTP(healthRec, healthReq)
	if healthRec.Code != http.StatusOK {
		t.Fatalf("expected health status 200, got %d", healthRec.Code)
	}

	metricsReq := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	metricsRec := httptest.NewRecorder()
	router.ServeHTTP(metricsRec, metricsReq)
	if metricsRec.Code != http.StatusOK {
		t.Fatalf("expected metrics status 200, got %d", metricsRec.Code)
	}

	body := metricsRec.Body.String()
	if !hasMetricLine(body, "http_requests_total", []string{`route="/health"`, `method="GET"`, `status="200"`}) {
		t.Fatalf("expected http_requests_total label set for /health")
	}
	if !hasMetricLine(body, "grpc_requests_total", []string{`method="/amss.v1.TaskService/GetTask"`, `code="OK"`}) {
		t.Fatalf("expected grpc_requests_total label set for gRPC method")
	}
}

func hasMetricLine(body, name string, labels []string) bool {
	for _, line := range strings.Split(body, "\n") {
		if !strings.HasPrefix(line, name+"{") {
			continue
		}
		matches := true
		for _, label := range labels {
			if !strings.Contains(line, label) {
				matches = false
				break
			}
		}
		if matches {
			return true
		}
	}
	return false
}
