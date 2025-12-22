package grpcapi

import (
	"context"

	amssv1 "github.com/aeromaintain/amss/api/proto/amssv1"
	"github.com/aeromaintain/amss/internal/app/services"
	"github.com/aeromaintain/amss/internal/domain"
)

type ProgramServiceServer struct {
	amssv1.UnimplementedProgramServiceServer
	Programs *services.MaintenanceProgramService
}

func (s *ProgramServiceServer) GenerateTasks(ctx context.Context, req *amssv1.GenerateTasksRequest) (*amssv1.GenerateTasksResponse, error) {
	if s.Programs == nil {
		return nil, mapError(domain.NewValidationError("program service unavailable"))
	}
	actor, err := actorFromContext(ctx)
	if err != nil {
		return nil, err
	}
	if actor.Role != domain.RoleAdmin {
		return nil, mapError(domain.ErrForbidden)
	}
	created, err := s.Programs.GenerateDueTasks(ctx, actor, 100)
	if err != nil {
		return nil, mapError(err)
	}
	return &amssv1.GenerateTasksResponse{Created: int32(created)}, nil
}
