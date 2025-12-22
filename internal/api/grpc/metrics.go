package grpcapi

import (
	"context"
	"time"

	"github.com/aeromaintain/amss/pkg/observability"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

func metricsUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		observability.ObserveGRPC(info.FullMethod, status.Code(err), time.Since(start))
		return resp, err
	}
}
