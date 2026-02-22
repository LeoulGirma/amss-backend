package services

import (
	"context"
	"fmt"
	"time"

	"github.com/aeromaintain/amss/internal/app"
	"github.com/aeromaintain/amss/internal/app/ports"
	"github.com/aeromaintain/amss/internal/domain"
	"github.com/google/uuid"
)

type PartReservationService struct {
	Reservations    ports.PartReservationRepository
	PartItems       ports.PartItemRepository
	PartDefinitions ports.PartDefinitionRepository
	Tasks           ports.TaskRepository
	Alerts          ports.AlertRepository
	Locker          ports.Locker
	Audit           ports.AuditRepository
	Outbox          ports.OutboxRepository
	Clock           app.Clock
}

func (s *PartReservationService) ListByTask(ctx context.Context, orgID, taskID uuid.UUID) ([]domain.PartReservation, error) {
	if s.Reservations == nil {
		return nil, nil
	}
	return s.Reservations.ListByTask(ctx, orgID, taskID)
}

func (s *PartReservationService) Reserve(ctx context.Context, actor app.Actor, taskID, partItemID uuid.UUID) (domain.PartReservation, error) {
	if s.Clock == nil {
		s.Clock = app.RealClock{}
	}
	if actor.Role != domain.RoleScheduler && actor.Role != domain.RoleMechanic && actor.Role != domain.RoleAdmin {
		return domain.PartReservation{}, domain.ErrForbidden
	}
	if s.Locker == nil {
		return domain.PartReservation{}, domain.NewConflictError("reservation lock unavailable")
	}
	lock, err := s.Locker.Acquire(ctx, fmt.Sprintf("part_item:%s", partItemID), 5*time.Second)
	if err != nil {
		return domain.PartReservation{}, err
	}
	defer func() { _ = lock.Release(ctx) }()

	item, err := s.PartItems.GetByID(ctx, actor.OrgID, partItemID)
	if err != nil {
		return domain.PartReservation{}, err
	}
	if item.Status != domain.PartItemInStock {
		return domain.PartReservation{}, domain.NewConflictError("part item not available")
	}

	reservation := domain.PartReservation{
		ID:         uuid.New(),
		OrgID:      actor.OrgID,
		TaskID:     taskID,
		PartItemID: partItemID,
		State:      domain.ReservationReserved,
		Quantity:   1,
		CreatedAt:  s.Clock.Now(),
		UpdatedAt:  s.Clock.Now(),
	}

	if err := s.Reservations.Create(ctx, reservation); err != nil {
		return domain.PartReservation{}, err
	}

	s.emitReservationAudit(ctx, actor, reservation, domain.AuditActionCreate)
	s.emitReservationOutbox(ctx, reservation, "part_reserved")

	return reservation, nil
}

func (s *PartReservationService) UpdateState(ctx context.Context, actor app.Actor, reservationID uuid.UUID, newState domain.PartReservationState) (domain.PartReservation, error) {
	if s.Clock == nil {
		s.Clock = app.RealClock{}
	}
	reservation, err := s.Reservations.GetByID(ctx, actor.OrgID, reservationID)
	if err != nil {
		return domain.PartReservation{}, err
	}

	if err := reservation.CanTransition(newState); err != nil {
		return domain.PartReservation{}, err
	}

	task, err := s.Tasks.GetByID(ctx, actor.OrgID, reservation.TaskID)
	if err != nil {
		return domain.PartReservation{}, err
	}
	if !actor.IsAdmin() && actor.OrgID != task.OrgID {
		return domain.PartReservation{}, domain.ErrForbidden
	}

	if newState == domain.ReservationUsed {
		if actor.Role == domain.RoleMechanic {
			if task.AssignedMechanicID == nil || *task.AssignedMechanicID != actor.UserID {
				return domain.PartReservation{}, domain.ErrForbidden
			}
		} else if actor.Role != domain.RoleScheduler && actor.Role != domain.RoleAdmin {
			return domain.PartReservation{}, domain.ErrForbidden
		}
	} else if newState == domain.ReservationReleased {
		if actor.Role != domain.RoleScheduler && actor.Role != domain.RoleMechanic && actor.Role != domain.RoleAdmin {
			return domain.PartReservation{}, domain.ErrForbidden
		}
	}

	if err := s.Reservations.UpdateState(ctx, actor.OrgID, reservation.ID, newState, s.Clock.Now()); err != nil {
		return domain.PartReservation{}, err
	}

	if newState == domain.ReservationUsed {
		if err := s.PartItems.UpdateStatus(ctx, actor.OrgID, reservation.PartItemID, domain.PartItemUsed, s.Clock.Now()); err != nil {
			return domain.PartReservation{}, err
		}
		// Check stock level and create alert if below threshold
		s.checkStockLevel(ctx, actor.OrgID, reservation.PartItemID)
	}

	reservation.State = newState
	reservation.UpdatedAt = s.Clock.Now()

	action := domain.AuditActionUpdate
	s.emitReservationAudit(ctx, actor, reservation, action)
	if newState == domain.ReservationUsed {
		s.emitReservationOutbox(ctx, reservation, "part_used")
	} else {
		s.emitReservationOutbox(ctx, reservation, "part_released")
	}

	return reservation, nil
}

// checkStockLevel checks if a part definition's stock has dropped below
// min_stock_level after a part item is used. If so, it creates a low-stock alert.
func (s *PartReservationService) checkStockLevel(ctx context.Context, orgID uuid.UUID, partItemID uuid.UUID) {
	if s.Alerts == nil || s.PartDefinitions == nil || s.PartItems == nil {
		return
	}

	// Get the part item to find its definition
	item, err := s.PartItems.GetByID(ctx, orgID, partItemID)
	if err != nil {
		return
	}

	// Get the definition for stock thresholds
	def, err := s.PartDefinitions.GetByID(ctx, orgID, item.DefinitionID)
	if err != nil {
		return
	}

	if def.MinStockLevel <= 0 {
		return // no threshold configured
	}

	// Count in-stock items for this definition
	inStock := domain.PartItemInStock
	items, err := s.PartItems.List(ctx, ports.PartItemFilter{
		OrgID:        &orgID,
		DefinitionID: &item.DefinitionID,
		Status:       &inStock,
	})
	if err != nil {
		return
	}

	currentStock := len(items)
	if currentStock >= def.MinStockLevel {
		return // stock is fine
	}

	// Create low-stock alert
	thresholdVal := float64(def.MinStockLevel)
	currentVal := float64(currentStock)
	alert := domain.Alert{
		ID:              uuid.New(),
		OrgID:           orgID,
		Level:           domain.AlertWarning,
		Category:        "parts_low_stock",
		Title:           fmt.Sprintf("Low stock: %s", def.Name),
		Description:     fmt.Sprintf("Stock level (%d) is below minimum (%d)", currentStock, def.MinStockLevel),
		EntityType:      "part_definition",
		EntityID:        def.ID,
		ThresholdValue:  &thresholdVal,
		CurrentValue:    &currentVal,
		CreatedAt:       s.Clock.Now(),
	}

	if currentStock == 0 {
		alert.Level = domain.AlertCritical
		alert.Title = fmt.Sprintf("Out of stock: %s", def.Name)
		// Auto-escalate critical alerts in 1 hour
		escalateAt := s.Clock.Now().Add(time.Hour)
		alert.AutoEscalateAt = &escalateAt
	}

	_, _ = s.Alerts.Create(ctx, alert)
}

func (s *PartReservationService) emitReservationAudit(ctx context.Context, actor app.Actor, reservation domain.PartReservation, action domain.AuditAction) {
	if s.Audit == nil {
		return
	}
	entry := domain.AuditLog{
		ID:            uuid.New(),
		OrgID:         reservation.OrgID,
		EntityType:    "part_reservation",
		EntityID:      reservation.ID,
		Action:        action,
		UserID:        actor.UserID,
		RequestID:     uuid.Nil,
		EntityVersion: 0,
		Timestamp:     s.Clock.Now(),
		Details: map[string]any{
			"state": reservation.State,
		},
	}
	_ = s.Audit.Insert(ctx, entry)
}

func (s *PartReservationService) emitReservationOutbox(ctx context.Context, reservation domain.PartReservation, eventType string) {
	if s.Outbox == nil {
		return
	}
	dedupeKey := fmt.Sprintf("%s:%s:%s", eventType, reservation.OrgID, reservation.ID)
	payload := map[string]any{
		"version":        1,
		"org_id":         reservation.OrgID,
		"reservation_id": reservation.ID,
		"task_id":        reservation.TaskID,
		"part_item_id":   reservation.PartItemID,
		"state":          reservation.State,
		"timestamp":      s.Clock.Now(),
	}
	_ = s.Outbox.Enqueue(ctx, reservation.OrgID, eventType, "part_reservation", reservation.ID, payload, dedupeKey)
}
