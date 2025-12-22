package rest

import (
	"context"
	"crypto/rsa"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"

	"github.com/aeromaintain/amss/internal/api/rest/handlers"
	amiddleware "github.com/aeromaintain/amss/internal/api/rest/middleware"
	"github.com/aeromaintain/amss/internal/app/services"
	postgresinfra "github.com/aeromaintain/amss/internal/infra/postgres"
	redisinfra "github.com/aeromaintain/amss/internal/infra/redis"
	"github.com/aeromaintain/amss/internal/infra/streams"
)

type Deps struct {
	Logger             zerolog.Logger
	DB                 *pgxpool.Pool
	Redis              *redis.Client
	PrometheusEnabled  bool
	AuthPublicKey      *rsa.PublicKey
	AuthPrivateKey     *rsa.PrivateKey
	AccessTokenTTL     time.Duration
	RefreshTokenTTL    time.Duration
	AppEnv             string
	ImportStorageDir   string
	CorsAllowedOrigins []string
}

func NewRouter(deps Deps) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	if len(deps.CorsAllowedOrigins) > 0 {
		r.Use(cors.Handler(cors.Options{
			AllowedOrigins:   deps.CorsAllowedOrigins,
			AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "Idempotency-Key", "X-API-Key", "X-Request-ID"},
			ExposedHeaders:   []string{"X-RateLimit-Remaining", "X-RateLimit-Reset", "Retry-After"},
			AllowCredentials: false,
			MaxAge:           300,
		}))
	}
	r.Use(amiddleware.RequestID)
	r.Use(amiddleware.Metrics())
	r.Use(amiddleware.Logging(deps.Logger))

	r.Route("/api/v1", func(api chi.Router) {
		idempotencyStore := &postgresinfra.IdempotencyStore{DB: deps.DB}
		rateLimiter := &redisinfra.RateLimiter{Client: deps.Redis}
		locker := &redisinfra.Locker{Client: deps.Redis}
		auditRepo := &postgresinfra.AuditRepository{DB: deps.DB}
		outboxRepo := &postgresinfra.OutboxRepository{DB: deps.DB}
		orgRepo := &postgresinfra.OrganizationRepository{DB: deps.DB}
		userRepo := &postgresinfra.UserRepository{DB: deps.DB}
		aircraftRepo := &postgresinfra.AircraftRepository{DB: deps.DB}
		programRepo := &postgresinfra.MaintenanceProgramRepository{DB: deps.DB}
		importRepo := &postgresinfra.ImportRepository{DB: deps.DB}
		importRowRepo := &postgresinfra.ImportRowRepository{DB: deps.DB}
		webhookRepo := &postgresinfra.WebhookRepository{DB: deps.DB}
		policyRepo := &postgresinfra.OrgPolicyRepository{DB: deps.DB}
		taskService := &services.TaskService{
			Tasks:        &postgresinfra.TaskRepository{DB: deps.DB},
			Aircraft:     aircraftRepo,
			Reservations: &postgresinfra.PartReservationRepository{DB: deps.DB},
			Compliance:   &postgresinfra.ComplianceRepository{DB: deps.DB},
			Audit:        auditRepo,
			Outbox:       outboxRepo,
		}
		partService := &services.PartReservationService{
			Reservations: &postgresinfra.PartReservationRepository{DB: deps.DB},
			PartItems:    &postgresinfra.PartItemRepository{DB: deps.DB},
			Tasks:        &postgresinfra.TaskRepository{DB: deps.DB},
			Locker:       locker,
			Audit:        auditRepo,
			Outbox:       outboxRepo,
		}
		complianceService := &services.ComplianceService{
			Compliance: &postgresinfra.ComplianceRepository{DB: deps.DB},
			Audit:      auditRepo,
			Outbox:     outboxRepo,
		}
		auditQueryService := &services.AuditQueryService{
			Repo: &postgresinfra.AuditQueryRepository{DB: deps.DB},
		}
		catalogService := &services.PartCatalogService{
			Definitions: &postgresinfra.PartDefinitionRepository{DB: deps.DB},
			Items:       &postgresinfra.PartItemRepository{DB: deps.DB},
		}
		orgService := &services.OrganizationService{
			Organizations: orgRepo,
		}
		userService := &services.UserService{
			Users: userRepo,
		}
		aircraftService := &services.AircraftService{
			Aircraft: aircraftRepo,
		}
		programService := &services.MaintenanceProgramService{
			Programs: programRepo,
			Tasks:    taskService.Tasks,
			TaskSvc:  taskService,
		}
		importService := &services.ImportService{
			Imports: importRepo,
			Rows:    importRowRepo,
			Jobs:    &streams.ImportQueue{Client: deps.Redis},
		}
		webhookService := &services.WebhookService{
			Webhooks:     webhookRepo,
			Outbox:       outboxRepo,
			RequireHTTPS: deps.AppEnv == "production",
		}
		policyService := &services.OrgPolicyService{
			Policies: policyRepo,
		}
		reportService := &services.ReportService{
			Reports: &postgresinfra.ReportRepository{DB: deps.DB},
		}

		authService := &services.AuthService{
			Users:         &postgresinfra.AuthRepository{DB: deps.DB},
			RefreshTokens: &postgresinfra.RefreshTokenRepository{DB: deps.DB},
			PrivateKey:    deps.AuthPrivateKey,
			PublicKey:     deps.AuthPublicKey,
			AccessTTL:     deps.AccessTokenTTL,
			RefreshTTL:    deps.RefreshTokenTTL,
		}
		authHandler := &handlers.AuthHandler{
			Service:     authService,
			RateLimiter: rateLimiter,
		}

		api.Route("/auth", func(auth chi.Router) {
			auth.Post("/login", authHandler.Login)
			auth.Post("/refresh", authHandler.Refresh)
			auth.Post("/logout", authHandler.Logout)
		})

		api.Group(func(protected chi.Router) {
			if deps.AuthPublicKey != nil {
				protected.Use(amiddleware.Authenticator{PublicKey: deps.AuthPublicKey}.Middleware)
			}
			protected.Use(amiddleware.InjectServices(amiddleware.ServiceRegistry{
				Tasks:         taskService,
				Parts:         partService,
				Compliance:    complianceService,
				AuditQuery:    auditQueryService,
				Catalog:       catalogService,
				Organizations: orgService,
				Users:         userService,
				Aircraft:      aircraftService,
				Programs:      programService,
				Imports:       importService,
				Webhooks:      webhookService,
				Policies:      policyService,
				Reports:       reportService,
			}))
			protected.Use(amiddleware.Idempotency(amiddleware.IdempotencyConfig{Store: idempotencyStore}))
			protected.Use(amiddleware.RateLimit(amiddleware.RateLimitConfig{
				Limiter:      rateLimiter,
				Window:       time.Minute,
				Category:     "default",
				DefaultLimit: 100,
				LimitFor: func(ctx context.Context, orgID uuid.UUID) (int, error) {
					policy, err := policyService.Get(ctx, orgID)
					if err != nil {
						return 0, err
					}
					return policy.APIRateLimitPerMin, nil
				},
				APIKeyHeader:       "X-API-Key",
				APIKeyCategory:     "api_key",
				APIKeyDefaultLimit: 10,
				APIKeyLimitFor: func(ctx context.Context, orgID uuid.UUID) (int, error) {
					policy, err := policyService.Get(ctx, orgID)
					if err != nil {
						return 0, err
					}
					return policy.APIKeyRateLimitPerMin, nil
				},
			}))

			protected.Route("/maintenance-tasks", func(tasks chi.Router) {
				tasks.Post("/", handlers.CreateTask)
				tasks.Get("/", handlers.ListTasks)
				tasks.Get("/{id}", handlers.GetTask)
				tasks.Patch("/{id}", handlers.UpdateTask)
				tasks.Delete("/{id}", handlers.DeleteTask)
				tasks.Patch("/{id}/state", handlers.TransitionTaskState)
			})
			protected.Route("/organizations", func(orgs chi.Router) {
				orgs.Post("/", handlers.CreateOrganization)
				orgs.Get("/", handlers.ListOrganizations)
				orgs.Get("/{id}", handlers.GetOrganization)
				orgs.Patch("/{id}", handlers.UpdateOrganization)
			})
			protected.Route("/users", func(users chi.Router) {
				users.Post("/", handlers.CreateUser)
				users.Get("/", handlers.ListUsers)
				users.Get("/{id}", handlers.GetUser)
				users.Patch("/{id}", handlers.UpdateUser)
				users.Delete("/{id}", handlers.DeleteUser)
			})
			protected.Route("/aircraft", func(aircraft chi.Router) {
				aircraft.Post("/", handlers.CreateAircraft)
				aircraft.Get("/", handlers.ListAircraft)
				aircraft.Get("/{id}", handlers.GetAircraft)
				aircraft.Patch("/{id}", handlers.UpdateAircraft)
				aircraft.Delete("/{id}", handlers.DeleteAircraft)
			})
			protected.Route("/maintenance-programs", func(programs chi.Router) {
				programs.Post("/", handlers.CreateProgram)
				programs.Get("/", handlers.ListPrograms)
				programs.Get("/{id}", handlers.GetProgram)
				programs.Patch("/{id}", handlers.UpdateProgram)
				programs.Delete("/{id}", handlers.DeleteProgram)
			})
			protected.Route("/part-definitions", func(defs chi.Router) {
				defs.Post("/", handlers.CreatePartDefinition)
				defs.Get("/", handlers.ListPartDefinitions)
				defs.Patch("/{id}", handlers.UpdatePartDefinition)
				defs.Delete("/{id}", handlers.DeletePartDefinition)
			})
			protected.Route("/part-items", func(items chi.Router) {
				items.Post("/", handlers.CreatePartItem)
				items.Get("/", handlers.ListPartItems)
				items.Patch("/{id}", handlers.UpdatePartItem)
				items.Delete("/{id}", handlers.DeletePartItem)
			})
			protected.Route("/part-reservations", func(parts chi.Router) {
				parts.Post("/", handlers.ReservePart)
				parts.Patch("/{id}/state", handlers.UpdateReservationState)
			})
			protected.Route("/compliance-items", func(compliance chi.Router) {
				compliance.Get("/", handlers.ListComplianceItems)
				compliance.Post("/", handlers.CreateComplianceItem)
				compliance.Patch("/{id}", handlers.UpdateComplianceItem)
				compliance.Patch("/{id}/sign-off", handlers.SignOffComplianceItem)
			})
			importHandler := handlers.ImportHandler{
				Service:    importService,
				StorageDir: deps.ImportStorageDir,
			}
			protected.Route("/imports", func(imports chi.Router) {
				imports.Post("/csv", importHandler.CreateImport)
				imports.Get("/{id}", importHandler.GetImport)
				imports.Get("/{id}/rows", importHandler.ListImportRows)
			})
			protected.Route("/webhooks", func(webhooks chi.Router) {
				webhooks.Post("/", handlers.CreateWebhook)
				webhooks.Get("/", handlers.ListWebhooks)
				webhooks.Delete("/{id}", handlers.DeleteWebhook)
				webhooks.Post("/{id}/test", handlers.TestWebhook)
			})
			protected.Route("/reports", func(reports chi.Router) {
				reports.Get("/summary", handlers.GetReportSummary)
				reports.Get("/compliance", handlers.GetComplianceReport)
			})
			protected.Get("/audit-logs", handlers.ListAuditLogs)
			protected.Get("/audit-logs/export", handlers.ExportAuditLogs)
		})
	})

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	r.Get("/openapi.yaml", OpenAPISpecHandler)
	r.Get("/docs", SwaggerUIHandler)

	r.Get("/ready", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if deps.DB != nil {
			if err := deps.DB.Ping(ctx); err != nil {
				msg := "db not ready"
				if deps.AppEnv != "production" {
					msg = fmt.Sprintf("db not ready: %v", err)
				}
				http.Error(w, msg, http.StatusServiceUnavailable)
				return
			}
		}
		if deps.Redis != nil {
			if err := deps.Redis.Ping(ctx).Err(); err != nil {
				msg := "redis not ready"
				if deps.AppEnv != "production" {
					msg = fmt.Sprintf("redis not ready: %v", err)
				}
				http.Error(w, msg, http.StatusServiceUnavailable)
				return
			}
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ready"))
	})

	if deps.PrometheusEnabled {
		r.Handle("/metrics", promhttp.Handler())
	}

	return r
}
