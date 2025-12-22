package jobs

import (
	"context"
	"testing"
	"time"

	"github.com/aeromaintain/amss/internal/app/services"
	"github.com/aeromaintain/amss/internal/domain"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

func TestProgramGeneratorCreatesTasks(t *testing.T) {
	ctx := context.Background()
	orgID := uuid.New()
	aircraftID := uuid.New()
	program := domain.MaintenanceProgram{
		ID:            uuid.New(),
		OrgID:         orgID,
		AircraftID:    &aircraftID,
		Name:          "A-Check",
		IntervalType:  domain.ProgramIntervalCalendar,
		IntervalValue: 30,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}

	programRepo := newFakeProgramRepo()
	programRepo.due = []domain.MaintenanceProgram{program}
	taskRepo := newFakeTaskRepo()
	taskService := &services.TaskService{Tasks: taskRepo}
	programService := &services.MaintenanceProgramService{
		Programs: programRepo,
		Tasks:    taskRepo,
		TaskSvc:  taskService,
	}

	generator := &ProgramGenerator{
		Programs: programService,
		Logger:   zerolog.Nop(),
	}
	generator.process(ctx)

	if len(taskRepo.tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(taskRepo.tasks))
	}
	for _, task := range taskRepo.tasks {
		if task.OrgID != orgID {
			t.Fatalf("expected org id %s, got %s", orgID, task.OrgID)
		}
		if task.ProgramID == nil || *task.ProgramID != program.ID {
			t.Fatalf("expected program id %s, got %v", program.ID, task.ProgramID)
		}
		if task.AircraftID != aircraftID {
			t.Fatalf("expected aircraft id %s, got %s", aircraftID, task.AircraftID)
		}
	}
}
