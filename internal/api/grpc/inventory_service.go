package grpcapi

import (
	"context"

	amssv1 "github.com/aeromaintain/amss/api/proto/amssv1"
	"github.com/aeromaintain/amss/internal/app/services"
	"github.com/aeromaintain/amss/internal/domain"
	"github.com/google/uuid"
)

type InventoryServiceServer struct {
	amssv1.UnimplementedInventoryServiceServer
	Parts *services.PartReservationService
}

func (s *InventoryServiceServer) ReserveParts(ctx context.Context, req *amssv1.ReservePartsRequest) (*amssv1.PartReservationResponse, error) {
	if req == nil {
		return nil, invalidArgument("missing request")
	}
	if s.Parts == nil {
		return nil, mapError(domain.NewValidationError("part service unavailable"))
	}
	actor, err := actorFromContext(ctx)
	if err != nil {
		return nil, err
	}
	if req.OrgId != "" && !actor.IsAdmin() && req.OrgId != actor.OrgID.String() {
		return nil, mapError(domain.ErrForbidden)
	}
	orgID, err := resolveOrgID(actor, req.OrgId)
	if err != nil {
		return nil, invalidArgument("invalid org_id")
	}
	actor = actorForOrg(actor, orgID)
	taskID, err := uuid.Parse(req.TaskId)
	if err != nil {
		return nil, invalidArgument("invalid task_id")
	}
	partItemID, err := uuid.Parse(req.PartItemId)
	if err != nil {
		return nil, invalidArgument("invalid part_item_id")
	}
	reservation, err := s.Parts.Reserve(ctx, actor, taskID, partItemID)
	if err != nil {
		return nil, mapError(err)
	}
	return &amssv1.PartReservationResponse{
		ReservationId: reservation.ID.String(),
		State:         string(reservation.State),
	}, nil
}

func (s *InventoryServiceServer) ReleaseParts(ctx context.Context, req *amssv1.ReleasePartsRequest) (*amssv1.PartReservationResponse, error) {
	if req == nil {
		return nil, invalidArgument("missing request")
	}
	if s.Parts == nil {
		return nil, mapError(domain.NewValidationError("part service unavailable"))
	}
	actor, err := actorFromContext(ctx)
	if err != nil {
		return nil, err
	}
	if req.OrgId != "" && !actor.IsAdmin() && req.OrgId != actor.OrgID.String() {
		return nil, mapError(domain.ErrForbidden)
	}
	orgID, err := resolveOrgID(actor, req.OrgId)
	if err != nil {
		return nil, invalidArgument("invalid org_id")
	}
	actor = actorForOrg(actor, orgID)
	reservationID, err := uuid.Parse(req.ReservationId)
	if err != nil {
		return nil, invalidArgument("invalid reservation_id")
	}
	reservation, err := s.Parts.UpdateState(ctx, actor, reservationID, domain.ReservationReleased)
	if err != nil {
		return nil, mapError(err)
	}
	return &amssv1.PartReservationResponse{
		ReservationId: reservation.ID.String(),
		State:         string(reservation.State),
	}, nil
}
