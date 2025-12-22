package jobs

import (
	"context"
	"encoding/csv"
	"errors"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aeromaintain/amss/internal/app/ports"
	"github.com/aeromaintain/amss/internal/domain"
	"github.com/aeromaintain/amss/pkg/observability"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

const (
	importConsumerGroup = "amss.imports"
	importJobsStream    = "import.jobs"
)

type ImportProcessor struct {
	Redis       *redis.Client
	Imports     ports.ImportRepository
	ImportRows  ports.ImportRowRepository
	Aircraft    ports.AircraftRepository
	Definitions ports.PartDefinitionRepository
	Items       ports.PartItemRepository
	Programs    ports.MaintenanceProgramRepository
	Logger      zerolog.Logger
	WorkerID    string
}

func (p *ImportProcessor) Run(ctx context.Context) {
	if p.Redis == nil {
		return
	}
	_ = p.Redis.XGroupCreateMkStream(ctx, importJobsStream, importConsumerGroup, "0").Err()
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		streams, err := p.Redis.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    importConsumerGroup,
			Consumer: p.WorkerID,
			Streams:  []string{importJobsStream, ">"},
			Count:    1,
			Block:    5 * time.Second,
		}).Result()
		if err != nil && !errors.Is(err, redis.Nil) {
			p.Logger.Error().Err(err).Msg("import stream read failed")
			continue
		}
		for _, stream := range streams {
			for _, message := range stream.Messages {
				p.handleMessage(ctx, message)
				_, _ = p.Redis.XAck(ctx, importJobsStream, importConsumerGroup, message.ID).Result()
			}
		}
	}
}

func (p *ImportProcessor) handleMessage(ctx context.Context, msg redis.XMessage) {
	idRaw, ok := msg.Values["import_id"]
	if !ok {
		return
	}
	var idStr string
	switch value := idRaw.(type) {
	case string:
		idStr = value
	case []byte:
		idStr = string(value)
	default:
		return
	}
	importID, err := uuid.Parse(idStr)
	if err != nil {
		return
	}
	p.processImport(ctx, importID)
}

func (p *ImportProcessor) processImport(ctx context.Context, importID uuid.UUID) {
	if p.Imports == nil || p.ImportRows == nil {
		return
	}
	observability.IncJobRun("import_processor")
	imp, err := p.Imports.GetByID(ctx, uuid.Nil, importID)
	if err != nil {
		observability.IncJobFailure("import_processor")
		p.Logger.Error().Err(err).Str("import_id", importID.String()).Msg("import not found")
		return
	}
	if imp.Status == domain.ImportStatusCompleted {
		return
	}

	_ = p.Imports.UpdateStatus(ctx, imp.OrgID, imp.ID, domain.ImportStatusValidating, nil, time.Now().UTC())

	file, err := os.Open(imp.FilePath)
	if err != nil {
		observability.IncJobFailure("import_processor")
		p.Imports.UpdateStatus(ctx, imp.OrgID, imp.ID, domain.ImportStatusFailed, map[string]any{"error": "file not found"}, time.Now().UTC())
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.TrimLeadingSpace = true
	header, err := reader.Read()
	if err != nil {
		observability.IncJobFailure("import_processor")
		p.Imports.UpdateStatus(ctx, imp.OrgID, imp.ID, domain.ImportStatusFailed, map[string]any{"error": "invalid header"}, time.Now().UTC())
		return
	}
	columns := normalizeColumns(header)

	var rows []domain.ImportRow
	rowNumber := 1
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		rowNumber++
		if err != nil {
			continue
		}
		rowData := mapColumns(columns, record)
		row := domain.ImportRow{
			ID:        uuid.New(),
			OrgID:     imp.OrgID,
			ImportID:  imp.ID,
			RowNumber: rowNumber,
			Raw:       toAnyMap(rowData),
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}
		errorsList := validateRow(imp.Type, rowData)
		if len(errorsList) > 0 {
			row.Status = domain.ImportRowInvalid
			row.Errors = errorsList
		} else {
			row.Status = domain.ImportRowValid
		}
		_ = p.ImportRows.Create(ctx, row)
		rows = append(rows, row)
	}

	_ = p.Imports.UpdateStatus(ctx, imp.OrgID, imp.ID, domain.ImportStatusApplying, nil, time.Now().UTC())

	applied := 0
	invalid := 0
	for _, row := range rows {
		if row.Status != domain.ImportRowValid {
			invalid++
			continue
		}
		err := p.applyRow(ctx, imp, row)
		if err != nil {
			row.Status = domain.ImportRowInvalid
			row.Errors = []string{err.Error()}
			row.UpdatedAt = time.Now().UTC()
			_ = p.ImportRows.Update(ctx, row)
			invalid++
			continue
		}
		row.Status = domain.ImportRowApplied
		row.UpdatedAt = time.Now().UTC()
		_ = p.ImportRows.Update(ctx, row)
		applied++
	}

	summary := map[string]any{
		"total":   len(rows),
		"applied": applied,
		"invalid": invalid,
	}
	status := domain.ImportStatusCompleted
	if invalid > 0 && applied == 0 {
		status = domain.ImportStatusFailed
	}
	_ = p.Imports.UpdateStatus(ctx, imp.OrgID, imp.ID, status, summary, time.Now().UTC())
}

func (p *ImportProcessor) applyRow(ctx context.Context, imp domain.Import, row domain.ImportRow) error {
	switch imp.Type {
	case domain.ImportTypeAircraft:
		return p.applyAircraft(ctx, imp.OrgID, row)
	case domain.ImportTypeParts:
		return p.applyParts(ctx, imp.OrgID, row)
	case domain.ImportTypePrograms:
		return p.applyPrograms(ctx, imp.OrgID, row)
	default:
		return errors.New("unsupported import type")
	}
}

func (p *ImportProcessor) applyAircraft(ctx context.Context, orgID uuid.UUID, row domain.ImportRow) error {
	if p.Aircraft == nil {
		return errors.New("aircraft repo unavailable")
	}
	data := toStringMap(row.Raw)
	tailNumber := data["tail_number"]
	if tailNumber == "" {
		return errors.New("tail_number is required")
	}
	model := data["model"]
	status := domain.AircraftStatus(data["status"])
	if status == "" {
		status = domain.AircraftOperational
	}
	capacity, _ := strconv.Atoi(data["capacity_slots"])
	lastMaintenance := parseOptionalTime(data["last_maintenance"])
	nextDue := parseOptionalTime(data["next_due"])
	flightHours, _ := strconv.Atoi(data["flight_hours_total"])
	cycles, _ := strconv.Atoi(data["cycles_total"])

	existing, err := p.Aircraft.GetByTailNumber(ctx, orgID, tailNumber)
	if err != nil {
		aircraft := domain.Aircraft{
			ID:               uuid.New(),
			OrgID:            orgID,
			TailNumber:       tailNumber,
			Model:            model,
			Status:           status,
			CapacitySlots:    capacity,
			LastMaintenance:  lastMaintenance,
			NextDue:          nextDue,
			FlightHoursTotal: flightHours,
			CyclesTotal:      cycles,
			CreatedAt:        time.Now().UTC(),
			UpdatedAt:        time.Now().UTC(),
		}
		_, err = p.Aircraft.Create(ctx, aircraft)
		return err
	}
	if model != "" {
		existing.Model = model
	}
	existing.Status = status
	if capacity > 0 {
		existing.CapacitySlots = capacity
	}
	if lastMaintenance != nil {
		existing.LastMaintenance = lastMaintenance
	}
	if nextDue != nil {
		existing.NextDue = nextDue
	}
	if flightHours > 0 {
		existing.FlightHoursTotal = flightHours
	}
	if cycles > 0 {
		existing.CyclesTotal = cycles
	}
	existing.UpdatedAt = time.Now().UTC()
	_, err = p.Aircraft.Update(ctx, existing)
	return err
}

func (p *ImportProcessor) applyParts(ctx context.Context, orgID uuid.UUID, row domain.ImportRow) error {
	if p.Definitions == nil || p.Items == nil {
		return errors.New("parts repo unavailable")
	}
	data := toStringMap(row.Raw)
	defName := data["definition_name"]
	category := data["category"]
	serial := data["serial_number"]
	if defName == "" || category == "" || serial == "" {
		return errors.New("definition_name, category, and serial_number are required")
	}

	definition, err := findDefinitionByName(ctx, p.Definitions, orgID, defName)
	if err != nil {
		definition = domain.PartDefinition{
			ID:        uuid.New(),
			OrgID:     orgID,
			Name:      defName,
			Category:  category,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}
		definition, err = p.Definitions.Create(ctx, definition)
		if err != nil {
			return err
		}
	}

	status := domain.PartItemStatus(data["status"])
	if status == "" {
		status = domain.PartItemInStock
	}
	expiry := parseOptionalTime(data["expiry_date"])

	item, err := p.Items.GetBySerialNumber(ctx, orgID, serial)
	if err != nil {
		newItem := domain.PartItem{
			ID:           uuid.New(),
			OrgID:        orgID,
			DefinitionID: definition.ID,
			SerialNumber: serial,
			Status:       status,
			ExpiryDate:   expiry,
			CreatedAt:    time.Now().UTC(),
			UpdatedAt:    time.Now().UTC(),
		}
		_, err = p.Items.Create(ctx, newItem)
		return err
	}

	item.DefinitionID = definition.ID
	item.Status = status
	if expiry != nil {
		item.ExpiryDate = expiry
	}
	item.UpdatedAt = time.Now().UTC()
	_, err = p.Items.Update(ctx, item)
	return err
}

func (p *ImportProcessor) applyPrograms(ctx context.Context, orgID uuid.UUID, row domain.ImportRow) error {
	if p.Programs == nil {
		return errors.New("program repo unavailable")
	}
	data := toStringMap(row.Raw)
	name := data["name"]
	intervalType := domain.MaintenanceProgramIntervalType(data["interval_type"])
	intervalValue, _ := strconv.Atoi(data["interval_value"])
	if name == "" || intervalType == "" || intervalValue <= 0 {
		return errors.New("name, interval_type, and interval_value are required")
	}
	var aircraftID *uuid.UUID
	if data["aircraft_id"] != "" {
		parsed, err := uuid.Parse(data["aircraft_id"])
		if err != nil {
			return errors.New("invalid aircraft_id")
		}
		aircraftID = &parsed
	}
	lastPerformed := parseOptionalTime(data["last_performed"])

	existing, err := p.Programs.GetByName(ctx, orgID, name, aircraftID)
	if err != nil {
		program := domain.MaintenanceProgram{
			ID:            uuid.New(),
			OrgID:         orgID,
			AircraftID:    aircraftID,
			Name:          name,
			IntervalType:  intervalType,
			IntervalValue: intervalValue,
			LastPerformed: lastPerformed,
			CreatedAt:     time.Now().UTC(),
			UpdatedAt:     time.Now().UTC(),
		}
		_, err = p.Programs.Create(ctx, program)
		return err
	}
	existing.IntervalType = intervalType
	existing.IntervalValue = intervalValue
	if lastPerformed != nil {
		existing.LastPerformed = lastPerformed
	}
	existing.UpdatedAt = time.Now().UTC()
	_, err = p.Programs.Update(ctx, existing)
	return err
}

func validateRow(importType domain.ImportType, row map[string]string) []string {
	var errorsList []string
	switch importType {
	case domain.ImportTypeAircraft:
		if row["tail_number"] == "" {
			errorsList = append(errorsList, "tail_number is required")
		}
		if row["model"] == "" {
			errorsList = append(errorsList, "model is required")
		}
		if row["capacity_slots"] == "" {
			errorsList = append(errorsList, "capacity_slots is required")
		}
	case domain.ImportTypeParts:
		if row["definition_name"] == "" {
			errorsList = append(errorsList, "definition_name is required")
		}
		if row["category"] == "" {
			errorsList = append(errorsList, "category is required")
		}
		if row["serial_number"] == "" {
			errorsList = append(errorsList, "serial_number is required")
		}
	case domain.ImportTypePrograms:
		if row["name"] == "" {
			errorsList = append(errorsList, "name is required")
		}
		if row["interval_type"] == "" {
			errorsList = append(errorsList, "interval_type is required")
		}
		if row["interval_value"] == "" {
			errorsList = append(errorsList, "interval_value is required")
		}
	default:
		errorsList = append(errorsList, "unsupported import type")
	}
	return errorsList
}

func normalizeColumns(cols []string) []string {
	out := make([]string, 0, len(cols))
	for _, col := range cols {
		col = strings.TrimSpace(strings.ToLower(col))
		out = append(out, col)
	}
	return out
}

func mapColumns(columns []string, record []string) map[string]string {
	row := map[string]string{}
	for i, col := range columns {
		if i >= len(record) {
			break
		}
		row[col] = strings.TrimSpace(record[i])
	}
	return row
}

func toAnyMap(value map[string]string) map[string]any {
	out := map[string]any{}
	for k, v := range value {
		out[k] = v
	}
	return out
}

func toStringMap(raw map[string]any) map[string]string {
	out := map[string]string{}
	for k, v := range raw {
		if str, ok := v.(string); ok {
			out[k] = str
		}
	}
	return out
}

func parseOptionalTime(value string) *time.Time {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil
	}
	return &parsed
}

func findDefinitionByName(ctx context.Context, repo ports.PartDefinitionRepository, orgID uuid.UUID, name string) (domain.PartDefinition, error) {
	defs, err := repo.List(ctx, ports.PartDefinitionFilter{
		OrgID: &orgID,
		Name:  name,
		Limit: 1,
	})
	if err != nil || len(defs) == 0 {
		return domain.PartDefinition{}, domain.ErrNotFound
	}
	for _, def := range defs {
		if strings.EqualFold(def.Name, name) {
			return def, nil
		}
	}
	return domain.PartDefinition{}, domain.ErrNotFound
}
