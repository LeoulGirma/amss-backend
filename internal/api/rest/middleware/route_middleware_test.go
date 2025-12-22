package middleware

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/aeromaintain/amss/internal/app/ports"
	"github.com/aeromaintain/amss/internal/domain"
	"github.com/aeromaintain/amss/pkg/auth"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type fakeRateLimiter struct {
	mu    sync.Mutex
	calls map[string]int
}

func (f *fakeRateLimiter) Allow(_ context.Context, key string, limit int, window time.Duration) (bool, int, time.Time, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.calls == nil {
		f.calls = make(map[string]int)
	}
	f.calls[key]++
	count := f.calls[key]
	remaining := limit - count
	resetAt := time.Now().Add(window)
	return count <= limit, remaining, resetAt, nil
}

type memoryIdempotencyStore struct {
	mu      sync.Mutex
	records map[string]ports.IdempotencyRecord
}

func newMemoryIdempotencyStore() *memoryIdempotencyStore {
	return &memoryIdempotencyStore{records: make(map[string]ports.IdempotencyRecord)}
}

func (s *memoryIdempotencyStore) Get(_ context.Context, orgID uuid.UUID, key, endpoint string) (ports.IdempotencyRecord, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	record, ok := s.records[recordKey(orgID, key, endpoint)]
	return record, ok, nil
}

func (s *memoryIdempotencyStore) CreatePlaceholder(_ context.Context, record ports.IdempotencyRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.records[recordKey(record.OrgID, record.Key, record.Endpoint)] = record
	return nil
}

func (s *memoryIdempotencyStore) UpdateResponse(_ context.Context, orgID uuid.UUID, key, endpoint string, statusCode int, response []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	storeKey := recordKey(orgID, key, endpoint)
	record, ok := s.records[storeKey]
	if !ok {
		return nil
	}
	record.StatusCode = statusCode
	record.Response = response
	s.records[storeKey] = record
	return nil
}

func recordKey(orgID uuid.UUID, key, endpoint string) string {
	return orgID.String() + "|" + key + "|" + endpoint
}

func generateToken(t *testing.T, privateKey *rsa.PrivateKey, orgID uuid.UUID, role domain.Role) string {
	t.Helper()
	pair, err := auth.GenerateTokenPair(uuid.NewString(), orgID.String(), string(role), nil, time.Now().UTC(), time.Minute, time.Hour, privateKey)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	return pair.AccessToken
}

func newRSAKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate rsa key: %v", err)
	}
	return key
}

func TestAuthMiddlewareRejectsMissingToken(t *testing.T) {
	privateKey := newRSAKey(t)
	router := chi.NewRouter()
	router.Use(Authenticator{PublicKey: &privateKey.PublicKey}.Middleware)
	router.Get("/protected", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rr.Code)
	}
}

func TestAuthMiddlewareAllowsValidToken(t *testing.T) {
	privateKey := newRSAKey(t)
	orgID := uuid.New()
	token := generateToken(t, privateKey, orgID, domain.RoleScheduler)

	router := chi.NewRouter()
	router.Use(Authenticator{PublicKey: &privateKey.PublicKey}.Middleware)
	router.Get("/protected", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
}

func TestRateLimitMiddlewareBlocks(t *testing.T) {
	privateKey := newRSAKey(t)
	orgID := uuid.New()
	token := generateToken(t, privateKey, orgID, domain.RoleScheduler)
	limiter := &fakeRateLimiter{}

	router := chi.NewRouter()
	router.Use(Authenticator{PublicKey: &privateKey.PublicKey}.Middleware)
	router.Use(RateLimit(RateLimitConfig{
		Limiter:      limiter,
		Window:       time.Minute,
		Category:     "test",
		DefaultLimit: 1,
	}))
	router.Get("/limited", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/limited", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
	if rr.Header().Get("X-RateLimit-Remaining") == "" {
		t.Fatalf("expected rate limit headers to be set")
	}

	req = httptest.NewRequest(http.MethodGet, "/limited", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusTooManyRequests {
		t.Fatalf("expected status 429, got %d", rr.Code)
	}
}

func TestIdempotencyMiddlewareReplaysResponse(t *testing.T) {
	privateKey := newRSAKey(t)
	orgID := uuid.New()
	token := generateToken(t, privateKey, orgID, domain.RoleScheduler)
	store := newMemoryIdempotencyStore()

	var calls int
	router := chi.NewRouter()
	router.Use(Authenticator{PublicKey: &privateKey.PublicKey}.Middleware)
	router.Use(Idempotency(IdempotencyConfig{Store: store}))
	router.Post("/items", func(w http.ResponseWriter, _ *http.Request) {
		calls++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "created"})
	})

	body := []byte(`{"name":"widget"}`)
	req := httptest.NewRequest(http.MethodPost, "/items", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Idempotency-Key", "abc-123")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", rr.Code)
	}
	if calls != 1 {
		t.Fatalf("expected handler to be called once, got %d", calls)
	}

	req = httptest.NewRequest(http.MethodPost, "/items", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Idempotency-Key", "abc-123")
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", rr.Code)
	}
	if calls != 1 {
		t.Fatalf("expected handler to be called once, got %d", calls)
	}
}

func TestIdempotencyMiddlewareConflictingPayload(t *testing.T) {
	privateKey := newRSAKey(t)
	orgID := uuid.New()
	token := generateToken(t, privateKey, orgID, domain.RoleScheduler)
	store := newMemoryIdempotencyStore()

	router := chi.NewRouter()
	router.Use(Authenticator{PublicKey: &privateKey.PublicKey}.Middleware)
	router.Use(Idempotency(IdempotencyConfig{Store: store}))
	router.Post("/items", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})

	req := httptest.NewRequest(http.MethodPost, "/items", bytes.NewReader([]byte(`{"name":"widget"}`)))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Idempotency-Key", "abc-123")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	req = httptest.NewRequest(http.MethodPost, "/items", bytes.NewReader([]byte(`{"name":"gadget"}`)))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Idempotency-Key", "abc-123")
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusConflict {
		t.Fatalf("expected status 409, got %d", rr.Code)
	}
}
