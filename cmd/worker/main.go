package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/aeromaintain/amss/internal/app/services"
	"github.com/aeromaintain/amss/internal/config"
	"github.com/aeromaintain/amss/internal/infra/postgres"
	"github.com/aeromaintain/amss/internal/jobs"
	"github.com/aeromaintain/amss/pkg/observability"
	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	logger := observability.SetupLogger(cfg.LogLevel)
	shutdownTracing, err := observability.SetupTracing(ctx, "amss-worker", cfg.OTLPEndpoint)
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

	auditRepo := &postgres.AuditRepository{DB: dbpool}
	outboxRepo := &postgres.OutboxRepository{DB: dbpool}
	webhookRepo := &postgres.WebhookRepository{DB: dbpool}
	webhookDeliveryRepo := &postgres.WebhookDeliveryRepository{DB: dbpool}
	policyRepo := &postgres.OrgPolicyRepository{DB: dbpool}
	aircraftRepo := &postgres.AircraftRepository{DB: dbpool}
	orgRepo := &postgres.OrganizationRepository{DB: dbpool}
	taskRepo := &postgres.TaskRepository{DB: dbpool}
	complianceRepo := &postgres.ComplianceRepository{DB: dbpool}
	reservationRepo := &postgres.PartReservationRepository{DB: dbpool}
	partItemRepo := &postgres.PartItemRepository{DB: dbpool}
	defRepo := &postgres.PartDefinitionRepository{DB: dbpool}
	programRepo := &postgres.MaintenanceProgramRepository{DB: dbpool}
	importRepo := &postgres.ImportRepository{DB: dbpool}
	importRowRepo := &postgres.ImportRowRepository{DB: dbpool}
	retentionRepo := &postgres.RetentionRepository{DB: dbpool}

	taskService := &services.TaskService{
		Tasks:        taskRepo,
		Aircraft:     aircraftRepo,
		Reservations: reservationRepo,
		Compliance:   complianceRepo,
		Audit:        auditRepo,
		Outbox:       outboxRepo,
	}
	programService := &services.MaintenanceProgramService{
		Programs: programRepo,
		Tasks:    taskRepo,
		TaskSvc:  taskService,
	}
	policyService := &services.OrgPolicyService{
		Policies: policyRepo,
	}

	outboxPublisher := &jobs.OutboxPublisher{
		Outbox:      outboxRepo,
		Webhooks:    webhookRepo,
		Deliveries:  webhookDeliveryRepo,
		Redis:       redisClient,
		Logger:      logger,
		WorkerID:    cfg.WorkerID,
		MaxAttempts: 10,
	}
	webhookDispatcher := &jobs.WebhookDispatcher{
		Deliveries: webhookDeliveryRepo,
		Webhooks:   webhookRepo,
		Outbox:     outboxRepo,
		Policies:   policyService,
		Logger:     logger,
	}
	importProcessor := &jobs.ImportProcessor{
		Redis:       redisClient,
		Imports:     importRepo,
		ImportRows:  importRowRepo,
		Aircraft:    aircraftRepo,
		Definitions: defRepo,
		Items:       partItemRepo,
		Programs:    programRepo,
		Logger:      logger,
		WorkerID:    cfg.WorkerID,
	}
	programGenerator := &jobs.ProgramGenerator{
		Programs: programService,
		Logger:   logger,
	}
	retentionCleaner := &jobs.RetentionCleaner{
		Orgs:      orgRepo,
		Retention: retentionRepo,
		Policies:  policyService,
		Logger:    logger,
		Interval:  cfg.RetentionCleanupInterval,
	}

	go outboxPublisher.Run(ctx)
	go webhookDispatcher.Run(ctx)
	go importProcessor.Run(ctx)
	go programGenerator.Run(ctx)
	go retentionCleaner.Run(ctx)

	logger.Info().Str("worker_id", cfg.WorkerID).Msg("worker started")
	<-ctx.Done()
	logger.Info().Msg("worker shutdown complete")
}
