package redis

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestRateLimiterIntegration(t *testing.T) {
	if os.Getenv("AMSS_INTEGRATION") != "1" {
		t.Skip("set AMSS_INTEGRATION=1 to run integration tests")
	}
	addr := os.Getenv("AMSS_TEST_REDIS_ADDR")
	var cleanup func()
	if addr == "" {
		var err error
		addr, cleanup, err = startRedisContainer(context.Background())
		if err != nil {
			t.Fatalf("start redis container: %v", err)
		}
		t.Cleanup(cleanup)
	}

	client := redis.NewClient(&redis.Options{Addr: addr})
	t.Cleanup(func() { _ = client.Close() })

	limiter := &RateLimiter{Client: client}
	ctx := context.Background()

	allowed, remaining, _, err := limiter.Allow(ctx, "rl:test", 1, time.Minute)
	if err != nil {
		t.Fatalf("allow first: %v", err)
	}
	if !allowed {
		t.Fatalf("expected first request allowed")
	}
	if remaining != 0 {
		t.Fatalf("expected remaining 0, got %d", remaining)
	}

	allowed, _, _, err = limiter.Allow(ctx, "rl:test", 1, time.Minute)
	if err != nil {
		t.Fatalf("allow second: %v", err)
	}
	if allowed {
		t.Fatalf("expected second request blocked")
	}
}

func startRedisContainer(ctx context.Context) (string, func(), error) {
	req := testcontainers.ContainerRequest{
		Image:        "redis:7",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForListeningPort("6379/tcp").WithStartupTimeout(30 * time.Second),
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return "", nil, err
	}
	host, err := container.Host(ctx)
	if err != nil {
		_ = container.Terminate(ctx)
		return "", nil, err
	}
	port, err := container.MappedPort(ctx, "6379/tcp")
	if err != nil {
		_ = container.Terminate(ctx)
		return "", nil, err
	}
	addr := fmt.Sprintf("%s:%s", host, port.Port())
	cleanup := func() {
		_ = container.Terminate(context.Background())
	}
	return addr, cleanup, nil
}
