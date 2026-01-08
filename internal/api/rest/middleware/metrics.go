package middleware

import (
	"bufio"
	"net"
	"net/http"
	"time"

	"github.com/aeromaintain/amss/pkg/observability"
	"github.com/go-chi/chi/v5"
)

type metricsWriter struct {
	http.ResponseWriter
	status int
}

func (w *metricsWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

// Hijack implements http.Hijacker interface for WebSocket support
func (w *metricsWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := w.ResponseWriter.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, http.ErrNotSupported
}

func Metrics() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			mw := &metricsWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(mw, r)
			route := chi.RouteContext(r.Context()).RoutePattern()
			if route == "" {
				route = r.URL.Path
			}
			observability.ObserveHTTPRequest(route, r.Method, mw.status, time.Since(start))
		})
	}
}
