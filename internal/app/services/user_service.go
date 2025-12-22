package services

import (
	"context"
	"strings"
	"time"

	"github.com/aeromaintain/amss/internal/app"
	"github.com/aeromaintain/amss/internal/app/ports"
	"github.com/aeromaintain/amss/internal/domain"
	"github.com/aeromaintain/amss/pkg/auth"
	"github.com/google/uuid"
)

type UserService struct {
	Users ports.UserRepository
	Clock app.Clock
}

type UserCreateInput struct {
	OrgID    *uuid.UUID
	Email    string
	Role     domain.Role
	Password string
}

type UserUpdateInput struct {
	Email    *string
	Role     *domain.Role
	Password *string
}

func (s *UserService) Create(ctx context.Context, actor app.Actor, input UserCreateInput) (domain.User, error) {
	if s.Clock == nil {
		s.Clock = app.RealClock{}
	}
	if actor.Role != domain.RoleAdmin && actor.Role != domain.RoleTenantAdmin {
		return domain.User{}, domain.ErrForbidden
	}
	if input.Email == "" || input.Password == "" {
		return domain.User{}, domain.NewValidationError("email and password are required")
	}
	role := input.Role
	if role == "" {
		return domain.User{}, domain.NewValidationError("role is required")
	}
	if role == domain.RoleAdmin && actor.Role != domain.RoleAdmin {
		return domain.User{}, domain.ErrForbidden
	}
	orgID := actor.OrgID
	if actor.Role == domain.RoleAdmin && input.OrgID != nil && *input.OrgID != uuid.Nil {
		orgID = *input.OrgID
	}
	if actor.Role == domain.RoleTenantAdmin && orgID != actor.OrgID {
		return domain.User{}, domain.ErrForbidden
	}

	hash, err := auth.HashPassword(input.Password)
	if err != nil {
		return domain.User{}, err
	}
	user := domain.User{
		ID:           uuid.New(),
		OrgID:        orgID,
		Email:        strings.TrimSpace(strings.ToLower(input.Email)),
		Role:         role,
		PasswordHash: hash,
		CreatedAt:    s.Clock.Now(),
		UpdatedAt:    s.Clock.Now(),
	}
	return s.Users.Create(ctx, user)
}

func (s *UserService) List(ctx context.Context, actor app.Actor, filter ports.UserFilter) ([]domain.User, error) {
	if actor.Role != domain.RoleAdmin && actor.Role != domain.RoleTenantAdmin {
		return nil, domain.ErrForbidden
	}
	if actor.Role != domain.RoleAdmin {
		filter.OrgID = &actor.OrgID
	}
	return s.Users.List(ctx, filter)
}

func (s *UserService) Get(ctx context.Context, actor app.Actor, orgID, id uuid.UUID) (domain.User, error) {
	if actor.Role != domain.RoleAdmin && actor.Role != domain.RoleTenantAdmin {
		return domain.User{}, domain.ErrForbidden
	}
	if actor.Role != domain.RoleAdmin && orgID != actor.OrgID {
		return domain.User{}, domain.ErrForbidden
	}
	return s.Users.GetByID(ctx, orgID, id)
}

func (s *UserService) Update(ctx context.Context, actor app.Actor, orgID, id uuid.UUID, input UserUpdateInput) (domain.User, error) {
	if s.Clock == nil {
		s.Clock = app.RealClock{}
	}
	if actor.Role != domain.RoleAdmin && actor.Role != domain.RoleTenantAdmin {
		return domain.User{}, domain.ErrForbidden
	}
	if actor.Role != domain.RoleAdmin && orgID != actor.OrgID {
		return domain.User{}, domain.ErrForbidden
	}
	user, err := s.Users.GetByID(ctx, orgID, id)
	if err != nil {
		return domain.User{}, err
	}
	if input.Email != nil {
		user.Email = strings.TrimSpace(strings.ToLower(*input.Email))
	}
	if input.Role != nil {
		if *input.Role == domain.RoleAdmin && actor.Role != domain.RoleAdmin {
			return domain.User{}, domain.ErrForbidden
		}
		user.Role = *input.Role
	}
	if input.Password != nil {
		hash, err := auth.HashPassword(*input.Password)
		if err != nil {
			return domain.User{}, err
		}
		user.PasswordHash = hash
	}
	user.UpdatedAt = s.Clock.Now()
	return s.Users.Update(ctx, user)
}

func (s *UserService) Delete(ctx context.Context, actor app.Actor, orgID, id uuid.UUID) error {
	if s.Clock == nil {
		s.Clock = app.RealClock{}
	}
	if actor.Role != domain.RoleAdmin && actor.Role != domain.RoleTenantAdmin {
		return domain.ErrForbidden
	}
	if actor.Role != domain.RoleAdmin && orgID != actor.OrgID {
		return domain.ErrForbidden
	}
	return s.Users.SoftDelete(ctx, orgID, id, s.Clock.Now())
}

func (s *UserService) TouchLastLogin(ctx context.Context, orgID, id uuid.UUID, now time.Time) error {
	if s.Users == nil {
		return nil
	}
	user, err := s.Users.GetByID(ctx, orgID, id)
	if err != nil {
		return err
	}
	user.LastLogin = &now
	user.UpdatedAt = now
	_, err = s.Users.Update(ctx, user)
	return err
}
