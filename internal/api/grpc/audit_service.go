package grpcapi

import (
	"context"
	"strings"

	amssv1 "github.com/aeromaintain/amss/api/proto/amssv1"
	"github.com/aeromaintain/amss/internal/app/services"
	"github.com/aeromaintain/amss/internal/domain"
	"github.com/google/uuid"
)

type AuditServiceServer struct {
	amssv1.UnimplementedAuditServiceServer
	Audit *services.AuditService
}

func (s *AuditServiceServer) LogAction(ctx context.Context, req *amssv1.AuditLogRequest) (*amssv1.AuditLogResponse, error) {
	if req == nil {
		return nil, invalidArgument("missing request")
	}
	if s.Audit == nil {
		return nil, mapError(domain.NewValidationError("audit service unavailable"))
	}
	actor, err := actorFromContext(ctx)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(req.EntityType) == "" {
		return nil, invalidArgument("entity_type is required")
	}
	if req.OrgId != "" && !actor.IsAdmin() && req.OrgId != actor.OrgID.String() {
		return nil, mapError(domain.ErrForbidden)
	}
	orgID, err := resolveOrgID(actor, req.OrgId)
	if err != nil {
		return nil, invalidArgument("invalid org_id")
	}
	entityID, err := uuid.Parse(req.EntityId)
	if err != nil {
		return nil, invalidArgument("invalid entity_id")
	}
	action := domain.AuditAction(req.Action)
	if !validAuditAction(action) {
		return nil, invalidArgument("invalid action")
	}
	var requestID *uuid.UUID
	if rid := RequestIDFromContext(ctx); rid != "" {
		if parsed, err := uuid.Parse(rid); err == nil {
			requestID = &parsed
		}
	}
	id, err := s.Audit.Log(ctx, actor, services.AuditLogInput{
		OrgID:      &orgID,
		EntityType: req.EntityType,
		EntityID:   entityID,
		Action:     action,
		RequestID:  requestID,
	})
	if err != nil {
		return nil, mapError(err)
	}
	return &amssv1.AuditLogResponse{AuditLogId: id.String()}, nil
}

func validAuditAction(action domain.AuditAction) bool {
	switch action {
	case domain.AuditActionCreate, domain.AuditActionUpdate, domain.AuditActionDelete, domain.AuditActionStateChange:
		return true
	default:
		return false
	}
}
