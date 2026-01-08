package middleware

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/aeromaintain/amss/internal/app/ports"
)

type IdempotencyConfig struct {
	Store        ports.IdempotencyStore
	TTL          time.Duration
	MaxBodyBytes int64
}

func Idempotency(cfg IdempotencyConfig) func(http.Handler) http.Handler {
	if cfg.TTL == 0 {
		cfg.TTL = 24 * time.Hour
	}
	if cfg.MaxBodyBytes == 0 {
		cfg.MaxBodyBytes = 1 << 20
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cfg.Store == nil {
				next.ServeHTTP(w, r)
				return
			}
			key := r.Header.Get("Idempotency-Key")
			if key == "" {
				next.ServeHTTP(w, r)
				return
			}
			principal, ok := PrincipalFromContext(r.Context())
			if !ok {
				writeError(w, r, http.StatusUnauthorized, "AUTH", "unauthorized")
				return
			}

			body, err := readBody(r, cfg.MaxBodyBytes)
			if err != nil {
				writeError(w, r, http.StatusBadRequest, "VALIDATION", "invalid request body")
				return
			}
			requestHash := hashRequest(r.Method, r.URL.Path, body)
			endpoint := r.Method + " " + r.URL.Path

			record, found, err := cfg.Store.Get(r.Context(), principal.OrgID, key, endpoint)
			if err != nil {
				writeError(w, r, http.StatusInternalServerError, "INTERNAL", "idempotency lookup failed")
				return
			}

			now := time.Now().UTC()
			if found && record.ExpiresAt.After(now) {
				if record.RequestHash != requestHash {
					writeError(w, r, http.StatusConflict, "CONFLICT", "idempotency key conflict")
					return
				}
				if record.StatusCode != 0 {
					writeStoredResponse(w, record.StatusCode, record.Response)
					return
				}
				writeError(w, r, http.StatusConflict, "CONFLICT", "request in progress")
				return
			}

			record = ports.IdempotencyRecord{
				OrgID:       principal.OrgID,
				Key:         key,
				Endpoint:    endpoint,
				RequestHash: requestHash,
				StatusCode:  0,
				CreatedAt:   now,
				ExpiresAt:   now.Add(cfg.TTL),
			}

			if err := cfg.Store.CreatePlaceholder(r.Context(), record); err != nil {
				writeError(w, r, http.StatusInternalServerError, "INTERNAL", "idempotency reservation failed")
				return
			}

			recorder := &responseRecorder{ResponseWriter: w, status: http.StatusOK, maxBytes: cfg.MaxBodyBytes}
			next.ServeHTTP(recorder, r)

			responseBody := recorder.body.Bytes()
			if recorder.overflow {
				responseBody = []byte("null")
			}
			_ = cfg.Store.UpdateResponse(r.Context(), principal.OrgID, key, endpoint, recorder.status, responseBody)
		})
	}
}

type responseRecorder struct {
	http.ResponseWriter
	status   int
	body     bytes.Buffer
	maxBytes int64
	overflow bool
}

func (r *responseRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *responseRecorder) Write(p []byte) (int, error) {
	if !r.overflow {
		if int64(r.body.Len()+len(p)) > r.maxBytes {
			r.overflow = true
		} else {
			r.body.Write(p)
		}
	}
	return r.ResponseWriter.Write(p)
}

// Hijack implements http.Hijacker interface for WebSocket support
func (r *responseRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := r.ResponseWriter.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, http.ErrNotSupported
}

func readBody(r *http.Request, maxBytes int64) ([]byte, error) {
	if r.Body == nil {
		return nil, nil
	}
	data, err := io.ReadAll(io.LimitReader(r.Body, maxBytes))
	if err != nil {
		return nil, err
	}
	r.Body = io.NopCloser(bytes.NewReader(data))
	return data, nil
}

func hashRequest(method, path string, body []byte) string {
	canonical := body
	if len(body) > 0 {
		var decoded any
		if err := json.Unmarshal(body, &decoded); err == nil {
			if normalized, err := json.Marshal(decoded); err == nil {
				canonical = normalized
			}
		}
	}
	h := sha256.Sum256([]byte(method + "|" + path + "|" + string(canonical)))
	return hex.EncodeToString(h[:])
}

func writeStoredResponse(w http.ResponseWriter, statusCode int, body []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if len(body) > 0 {
		_, _ = w.Write(body)
	}
}
