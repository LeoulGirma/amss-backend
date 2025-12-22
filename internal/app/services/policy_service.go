package services

import (
	"context"
	"time"

	"github.com/aeromaintain/amss/internal/app"
	"github.com/aeromaintain/amss/internal/app/ports"
	"github.com/aeromaintain/amss/internal/domain"
	"github.com/google/uuid"
)

type OrgPolicyService struct {
	Policies ports.OrgPolicyRepository
	Clock    app.Clock
}

func (s *OrgPolicyService) Get(ctx context.Context, orgID uuid.UUID) (domain.OrgPolicy, error) {
	if s.Policies == nil {
		return defaultOrgPolicy(orgID, time.Now().UTC()), nil
	}
	policy, err := s.Policies.GetByOrgID(ctx, orgID)
	if err != nil {
		return defaultOrgPolicy(orgID, time.Now().UTC()), nil
	}
	return policy, nil
}

func (s *OrgPolicyService) Upsert(ctx context.Context, actor app.Actor, policy domain.OrgPolicy) (domain.OrgPolicy, error) {
	if s.Clock == nil {
		s.Clock = app.RealClock{}
	}
	if actor.Role != domain.RoleAdmin && actor.Role != domain.RoleTenantAdmin {
		return domain.OrgPolicy{}, domain.ErrForbidden
	}
	if !actor.IsAdmin() && policy.OrgID != actor.OrgID {
		return domain.OrgPolicy{}, domain.ErrForbidden
	}
	if policy.RetentionInterval <= 0 {
		policy.RetentionInterval = 365 * 24 * time.Hour
	}
	if policy.MaxWebhookAttempts <= 0 {
		policy.MaxWebhookAttempts = 10
	}
	if policy.WebhookReplayWindowSeconds <= 0 {
		policy.WebhookReplayWindowSeconds = 300
	}
	if policy.APIRateLimitPerMin <= 0 {
		policy.APIRateLimitPerMin = 100
	}
	if policy.APIKeyRateLimitPerMin <= 0 {
		policy.APIKeyRateLimitPerMin = 10
	}
	policy.UpdatedAt = s.Clock.Now()
	if policy.CreatedAt.IsZero() {
		policy.CreatedAt = policy.UpdatedAt
	}
	if s.Policies == nil {
		return policy, nil
	}
	return s.Policies.Upsert(ctx, policy)
}

func defaultOrgPolicy(orgID uuid.UUID, now time.Time) domain.OrgPolicy {
	return domain.OrgPolicy{
		OrgID:                      orgID,
		RetentionInterval:          365 * 24 * time.Hour,
		MaxWebhookAttempts:         10,
		WebhookReplayWindowSeconds: 300,
		APIRateLimitPerMin:         100,
		APIKeyRateLimitPerMin:      10,
		CreatedAt:                  now,
		UpdatedAt:                  now,
	}
}
