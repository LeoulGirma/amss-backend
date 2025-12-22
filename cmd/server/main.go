package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	grpcapi "github.com/aeromaintain/amss/internal/api/grpc"
	"github.com/aeromaintain/amss/internal/api/rest"
	"github.com/aeromaintain/amss/internal/app/services"
	"github.com/aeromaintain/amss/internal/config"
	postgresinfra "github.com/aeromaintain/amss/internal/infra/postgres"
	redisinfra "github.com/aeromaintain/amss/internal/infra/redis"
	"github.com/aeromaintain/amss/pkg/auth"
	"github.com/aeromaintain/amss/pkg/observability"
	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	logger := observability.SetupLogger(cfg.LogLevel)
	shutdownTracing, err := observability.SetupTracing(ctx, "amss-server", cfg.OTLPEndpoint)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to setup tracing")
	}
	defer func() { _ = shutdownTracing(context.Background()) }()

	observability.RegisterMetrics()

	dbConfig, err := pgxpool.ParseConfig(cfg.DBURL)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to parse database config")
	}
	dbConfig.ConnConfig.Tracer = otelpgx.NewTracer()
	dbpool, err := pgxpool.NewWithConfig(ctx, dbConfig)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to connect to database")
	}
	defer dbpool.Close()

	redisClient := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})
	defer func() { _ = redisClient.Close() }()

	privateKey, err := auth.ParseRSAPrivateKeyFromPEM(cfg.JWTPrivateKeyPEM)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to parse jwt private key")
	}
	publicKey, err := auth.ParseRSAPublicKeyFromPEM(cfg.JWTPublicKeyPEM)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to parse jwt public key")
	}

	restHandler := rest.NewRouter(rest.Deps{
		Logger:             logger,
		DB:                 dbpool,
		Redis:              redisClient,
		PrometheusEnabled:  cfg.PrometheusEnabled,
		AuthPublicKey:      publicKey,
		AuthPrivateKey:     privateKey,
		AccessTokenTTL:     cfg.AccessTokenTTL,
		RefreshTokenTTL:    cfg.RefreshTokenTTL,
		AppEnv:             cfg.AppEnv,
		ImportStorageDir:   cfg.ImportStorageDir,
		CorsAllowedOrigins: cfg.CorsAllowedOrigins,
	})
	restHandler = otelhttp.NewHandler(restHandler, "http-server")

	httpServer := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           restHandler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	taskService := &services.TaskService{
		Tasks:        &postgresinfra.TaskRepository{DB: dbpool},
		Aircraft:     &postgresinfra.AircraftRepository{DB: dbpool},
		Reservations: &postgresinfra.PartReservationRepository{DB: dbpool},
		Compliance:   &postgresinfra.ComplianceRepository{DB: dbpool},
		Audit:        &postgresinfra.AuditRepository{DB: dbpool},
		Outbox:       &postgresinfra.OutboxRepository{DB: dbpool},
	}
	partService := &services.PartReservationService{
		Reservations: &postgresinfra.PartReservationRepository{DB: dbpool},
		PartItems:    &postgresinfra.PartItemRepository{DB: dbpool},
		Tasks:        &postgresinfra.TaskRepository{DB: dbpool},
		Locker:       &redisinfra.Locker{Client: redisClient},
		Audit:        &postgresinfra.AuditRepository{DB: dbpool},
		Outbox:       &postgresinfra.OutboxRepository{DB: dbpool},
	}
	programService := &services.MaintenanceProgramService{
		Programs: &postgresinfra.MaintenanceProgramRepository{DB: dbpool},
		Tasks:    &postgresinfra.TaskRepository{DB: dbpool},
		TaskSvc:  taskService,
	}
	auditService := &services.AuditService{
		Repo: &postgresinfra.AuditRepository{DB: dbpool},
	}

	grpcServer := grpcapi.NewServer(grpcapi.Deps{
		Logger:        logger,
		AuthPublicKey: publicKey,
		Tasks:         taskService,
		Parts:         partService,
		Audit:         auditService,
		Programs:      programService,
	})
	grpcListener, err := net.Listen("tcp", cfg.GRPCAddr)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to listen on grpc addr")
	}

	go func() {
		logger.Info().Str("addr", cfg.GRPCAddr).Msg("grpc server starting")
		if err := grpcServer.Serve(grpcListener); err != nil {
			logger.Fatal().Err(err).Msg("grpc server failed")
		}
	}()

	go func() {
		logger.Info().Str("addr", cfg.HTTPAddr).Msg("http server starting")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal().Err(err).Msg("http server failed")
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	grpcServer.GracefulStop()
	_ = httpServer.Shutdown(shutdownCtx)
	logger.Info().Msg("server shutdown complete")
	_ = shutdownTracing(shutdownCtx)
}
