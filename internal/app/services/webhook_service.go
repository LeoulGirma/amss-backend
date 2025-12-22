package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/url"
	"strings"

	"github.com/aeromaintain/amss/internal/app"
	"github.com/aeromaintain/amss/internal/app/ports"
	"github.com/aeromaintain/amss/internal/domain"
	"github.com/google/uuid"
)

type WebhookService struct {
	Webhooks     ports.WebhookRepository
	Outbox       ports.OutboxRepository
	RequireHTTPS bool
	Clock        app.Clock
}

type WebhookCreateInput struct {
	OrgID  *uuid.UUID
	URL    string
	Events []string
}

func (s *WebhookService) Create(ctx context.Context, actor app.Actor, input WebhookCreateInput) (domain.Webhook, error) {
	if s.Clock == nil {
		s.Clock = app.RealClock{}
	}
	if actor.Role != domain.RoleTenantAdmin && actor.Role != domain.RoleAdmin {
		return domain.Webhook{}, domain.ErrForbidden
	}
	if input.URL == "" || len(input.Events) == 0 {
		return domain.Webhook{}, domain.NewValidationError("url and events are required")
	}
	if s.RequireHTTPS {
		parsed, err := url.Parse(input.URL)
		if err != nil || parsed.Scheme != "https" {
			return domain.Webhook{}, domain.NewValidationError("url must be https")
		}
	}
	orgID := actor.OrgID
	if actor.IsAdmin() && input.OrgID != nil && *input.OrgID != uuid.Nil {
		orgID = *input.OrgID
	}

	secret, err := generateSecret()
	if err != nil {
		return domain.Webhook{}, err
	}
	events := normalizeEventList(input.Events)
	hook := domain.Webhook{
		ID:        uuid.New(),
		OrgID:     orgID,
		URL:       strings.TrimSpace(input.URL),
		Events:    events,
		Secret:    secret,
		CreatedAt: s.Clock.Now(),
		UpdatedAt: s.Clock.Now(),
	}
	return s.Webhooks.Create(ctx, hook)
}

func (s *WebhookService) List(ctx context.Context, actor app.Actor, orgID uuid.UUID) ([]domain.Webhook, error) {
	if actor.Role != domain.RoleTenantAdmin && actor.Role != domain.RoleAdmin {
		return nil, domain.ErrForbidden
	}
	if !actor.IsAdmin() && orgID != actor.OrgID {
		return nil, domain.ErrForbidden
	}
	return s.Webhooks.List(ctx, orgID)
}

func (s *WebhookService) Delete(ctx context.Context, actor app.Actor, orgID, id uuid.UUID) error {
	if actor.Role != domain.RoleTenantAdmin && actor.Role != domain.RoleAdmin {
		return domain.ErrForbidden
	}
	if !actor.IsAdmin() && orgID != actor.OrgID {
		return domain.ErrForbidden
	}
	return s.Webhooks.Delete(ctx, orgID, id)
}

func (s *WebhookService) SendTest(ctx context.Context, actor app.Actor, orgID, id uuid.UUID) error {
	if actor.Role != domain.RoleTenantAdmin && actor.Role != domain.RoleAdmin {
		return domain.ErrForbidden
	}
	if !actor.IsAdmin() && orgID != actor.OrgID {
		return domain.ErrForbidden
	}
	if s.Outbox == nil {
		return domain.NewValidationError("outbox unavailable")
	}
	payload := map[string]any{
		"version":    1,
		"org_id":     orgID,
		"webhook_id": id,
		"timestamp":  s.Clock.Now(),
		"type":       "test",
	}
	return s.Outbox.Enqueue(ctx, orgID, "webhook.test", "webhook", id, payload, "webhook.test:"+orgID.String()+":"+id.String())
}

func normalizeEventList(events []string) []string {
	out := make([]string, 0, len(events))
	seen := map[string]struct{}{}
	for _, event := range events {
		event = strings.TrimSpace(event)
		if event == "" {
			continue
		}
		if _, ok := seen[event]; ok {
			continue
		}
		seen[event] = struct{}{}
		out = append(out, event)
	}
	return out
}

func generateSecret() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
