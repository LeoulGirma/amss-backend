package jobs

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/aeromaintain/amss/internal/domain"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

func TestImportProcessorProcessAircraft(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	path := filepath.Join(dir, "aircraft.csv")
	if err := os.WriteFile(path, []byte("tail_number,model,capacity_slots,status\nN123,737,10,operational\n"), 0o644); err != nil {
		t.Fatalf("write csv: %v", err)
	}

	orgID := uuid.New()
	importID := uuid.New()
	importRepo := newFakeImportRepo()
	rowRepo := newFakeImportRowRepo()
	aircraftRepo := newFakeAircraftRepo()

	_, _ = importRepo.Create(ctx, domain.Import{
		ID:        importID,
		OrgID:     orgID,
		Type:      domain.ImportTypeAircraft,
		Status:    domain.ImportStatusPending,
		FileName:  "aircraft.csv",
		FilePath:  path,
		CreatedBy: uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	})

	processor := &ImportProcessor{
		Imports:    importRepo,
		ImportRows: rowRepo,
		Aircraft:   aircraftRepo,
		Logger:     zerolog.Nop(),
	}
	processor.processImport(ctx, importID)

	updated, err := importRepo.GetByID(ctx, orgID, importID)
	if err != nil {
		t.Fatalf("fetch import: %v", err)
	}
	if updated.Status != domain.ImportStatusCompleted {
		t.Fatalf("expected status completed, got %s", updated.Status)
	}
	if len(aircraftRepo.aircraft) != 1 {
		t.Fatalf("expected 1 aircraft, got %d", len(aircraftRepo.aircraft))
	}
}

func TestImportProcessorProcessPrograms(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	aircraftID := uuid.New()
	path := filepath.Join(dir, "programs.csv")
	content := "name,interval_type,interval_value,aircraft_id\nA-Check,calendar,30," + aircraftID.String() + "\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write csv: %v", err)
	}

	orgID := uuid.New()
	importID := uuid.New()
	importRepo := newFakeImportRepo()
	rowRepo := newFakeImportRowRepo()
	programRepo := newFakeProgramRepo()

	_, _ = importRepo.Create(ctx, domain.Import{
		ID:        importID,
		OrgID:     orgID,
		Type:      domain.ImportTypePrograms,
		Status:    domain.ImportStatusPending,
		FileName:  "programs.csv",
		FilePath:  path,
		CreatedBy: uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	})

	processor := &ImportProcessor{
		Imports:    importRepo,
		ImportRows: rowRepo,
		Programs:   programRepo,
		Logger:     zerolog.Nop(),
	}
	processor.processImport(ctx, importID)

	updated, err := importRepo.GetByID(ctx, orgID, importID)
	if err != nil {
		t.Fatalf("fetch import: %v", err)
	}
	if updated.Status != domain.ImportStatusCompleted {
		t.Fatalf("expected status completed, got %s", updated.Status)
	}
	if len(programRepo.programs) != 1 {
		t.Fatalf("expected 1 program, got %d", len(programRepo.programs))
	}
	for _, program := range programRepo.programs {
		if program.AircraftID == nil || *program.AircraftID != aircraftID {
			t.Fatalf("expected aircraft id %s, got %v", aircraftID, program.AircraftID)
		}
	}
}

func TestImportProcessorMissingFile(t *testing.T) {
	ctx := context.Background()
	orgID := uuid.New()
	importID := uuid.New()
	importRepo := newFakeImportRepo()
	rowRepo := newFakeImportRowRepo()

	_, _ = importRepo.Create(ctx, domain.Import{
		ID:        importID,
		OrgID:     orgID,
		Type:      domain.ImportTypeAircraft,
		Status:    domain.ImportStatusPending,
		FileName:  "missing.csv",
		FilePath:  filepath.Join(t.TempDir(), "missing.csv"),
		CreatedBy: uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	})

	processor := &ImportProcessor{
		Imports:    importRepo,
		ImportRows: rowRepo,
		Logger:     zerolog.Nop(),
	}
	processor.processImport(ctx, importID)

	updated, err := importRepo.GetByID(ctx, orgID, importID)
	if err != nil {
		t.Fatalf("fetch import: %v", err)
	}
	if updated.Status != domain.ImportStatusFailed {
		t.Fatalf("expected status failed, got %s", updated.Status)
	}
}

func TestImportProcessorInvalidRows(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	path := filepath.Join(dir, "aircraft.csv")
	if err := os.WriteFile(path, []byte("tail_number,model,capacity_slots\nN123,737,\n"), 0o644); err != nil {
		t.Fatalf("write csv: %v", err)
	}

	orgID := uuid.New()
	importID := uuid.New()
	importRepo := newFakeImportRepo()
	rowRepo := newFakeImportRowRepo()
	aircraftRepo := newFakeAircraftRepo()

	_, _ = importRepo.Create(ctx, domain.Import{
		ID:        importID,
		OrgID:     orgID,
		Type:      domain.ImportTypeAircraft,
		Status:    domain.ImportStatusPending,
		FileName:  "aircraft.csv",
		FilePath:  path,
		CreatedBy: uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	})

	processor := &ImportProcessor{
		Imports:    importRepo,
		ImportRows: rowRepo,
		Aircraft:   aircraftRepo,
		Logger:     zerolog.Nop(),
	}
	processor.processImport(ctx, importID)

	updated, err := importRepo.GetByID(ctx, orgID, importID)
	if err != nil {
		t.Fatalf("fetch import: %v", err)
	}
	if updated.Status != domain.ImportStatusFailed {
		t.Fatalf("expected status failed, got %s", updated.Status)
	}
	if len(rowRepo.rows) != 1 {
		t.Fatalf("expected 1 import row, got %d", len(rowRepo.rows))
	}
	for _, row := range rowRepo.rows {
		if row.Status != domain.ImportRowInvalid {
			t.Fatalf("expected row status invalid, got %s", row.Status)
		}
	}
}

func TestImportProcessorUnsupportedType(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	path := filepath.Join(dir, "unknown.csv")
	if err := os.WriteFile(path, []byte("name\nvalue\n"), 0o644); err != nil {
		t.Fatalf("write csv: %v", err)
	}

	orgID := uuid.New()
	importID := uuid.New()
	importRepo := newFakeImportRepo()
	rowRepo := newFakeImportRowRepo()

	_, _ = importRepo.Create(ctx, domain.Import{
		ID:        importID,
		OrgID:     orgID,
		Type:      domain.ImportType("unknown"),
		Status:    domain.ImportStatusPending,
		FileName:  "unknown.csv",
		FilePath:  path,
		CreatedBy: uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	})

	processor := &ImportProcessor{
		Imports:    importRepo,
		ImportRows: rowRepo,
		Logger:     zerolog.Nop(),
	}
	processor.processImport(ctx, importID)

	updated, err := importRepo.GetByID(ctx, orgID, importID)
	if err != nil {
		t.Fatalf("fetch import: %v", err)
	}
	if updated.Status != domain.ImportStatusFailed {
		t.Fatalf("expected status failed, got %s", updated.Status)
	}
}
