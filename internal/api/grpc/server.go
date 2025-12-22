package grpcapi

import (
	"context"
	"crypto/rsa"
	"time"

	amssv1 "github.com/aeromaintain/amss/api/proto/amssv1"
	"github.com/aeromaintain/amss/internal/app/services"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type ctxKey int

const requestIDKey ctxKey = 0

type Deps struct {
	Logger        zerolog.Logger
	AuthPublicKey *rsa.PublicKey
	Tasks         *services.TaskService
	Parts         *services.PartReservationService
	Audit         *services.AuditService
	Programs      *services.MaintenanceProgramService
}

func NewServer(deps Deps) *grpc.Server {
	interceptors := []grpc.UnaryServerInterceptor{
		requestIDUnaryInterceptor(deps.Logger),
		metricsUnaryInterceptor(),
	}
	if deps.AuthPublicKey != nil {
		interceptors = append(interceptors, authUnaryInterceptor(deps.AuthPublicKey))
	}
	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(interceptors...),
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
	)
	if deps.Tasks != nil {
		amssv1.RegisterTaskServiceServer(server, &TaskServiceServer{Tasks: deps.Tasks})
	}
	if deps.Parts != nil {
		amssv1.RegisterInventoryServiceServer(server, &InventoryServiceServer{Parts: deps.Parts})
	}
	if deps.Audit != nil {
		amssv1.RegisterAuditServiceServer(server, &AuditServiceServer{Audit: deps.Audit})
	}
	if deps.Programs != nil {
		amssv1.RegisterProgramServiceServer(server, &ProgramServiceServer{Programs: deps.Programs})
	}
	return server
}

func requestIDUnaryInterceptor(logger zerolog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		md, _ := metadata.FromIncomingContext(ctx)
		rid := ""
		if md != nil {
			vals := md.Get("x-request-id")
			if len(vals) > 0 {
				rid = vals[0]
			}
		}
		if rid == "" {
			rid = uuid.NewString()
		}
		ctx = context.WithValue(ctx, requestIDKey, rid)
		start := time.Now()
		resp, err := handler(ctx, req)
		logger.Info().
			Str("request_id", rid).
			Str("method", info.FullMethod).
			Dur("duration", time.Since(start)).
			Err(err).
			Msg("grpc_request")
		return resp, err
	}
}

func RequestIDFromContext(ctx context.Context) string {
	if v := ctx.Value(requestIDKey); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
