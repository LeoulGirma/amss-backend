package postgres

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/aeromaintain/amss/internal/app/ports"
	"github.com/aeromaintain/amss/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DirectiveRepository struct {
	DB *pgxpool.Pool
}

// --- Regulatory Authorities ---

func (r *DirectiveRepository) ListAuthorities(ctx context.Context) ([]domain.RegulatoryAuthority, error) {
	rows, err := r.DB.Query(ctx, `
		SELECT id, code, name, country, record_retention_years, created_at
		FROM regulatory_authorities
		ORDER BY code
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.RegulatoryAuthority
	for rows.Next() {
		a, err := scanAuthority(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, a)
	}
	return items, rows.Err()
}

func (r *DirectiveRepository) GetAuthorityByID(ctx context.Context, id uuid.UUID) (domain.RegulatoryAuthority, error) {
	row := r.DB.QueryRow(ctx, `
		SELECT id, code, name, country, record_retention_years, created_at
		FROM regulatory_authorities
		WHERE id=$1
	`, id)
	return scanAuthority(row)
}

func scanAuthority(row pgx.Row) (domain.RegulatoryAuthority, error) {
	var a domain.RegulatoryAuthority
	if err := row.Scan(&a.ID, &a.Code, &a.Name, &a.Country, &a.RecordRetentionYears, &a.CreatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return domain.RegulatoryAuthority{}, domain.ErrNotFound
		}
		return domain.RegulatoryAuthority{}, err
	}
	return a, nil
}

// --- Compliance Directives ---

func (r *DirectiveRepository) GetDirectiveByID(ctx context.Context, id uuid.UUID) (domain.ComplianceDirective, error) {
	row := r.DB.QueryRow(ctx, `
		SELECT id, org_id, authority_id, directive_type, reference_number, title, description,
		       applicability, affected_aircraft_types, effective_date, compliance_deadline,
		       recurrence_interval, superseded_by, source_url, created_at, updated_at
		FROM compliance_directives
		WHERE id=$1
	`, id)
	return scanDirective(row)
}

func (r *DirectiveRepository) ListDirectives(ctx context.Context, filter ports.DirectiveFilter) ([]domain.ComplianceDirective, error) {
	clauses := make([]string, 0, 4)
	args := make([]any, 0, 6)
	add := func(condition string, value any) {
		args = append(args, value)
		clauses = append(clauses, condition+"$"+itoa(len(args)))
	}
	if filter.OrgID != nil {
		add("org_id=", *filter.OrgID)
	}
	if filter.AuthorityID != nil {
		add("authority_id=", *filter.AuthorityID)
	}
	if filter.DirectiveType != nil {
		add("directive_type=", *filter.DirectiveType)
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	query := `
		SELECT id, org_id, authority_id, directive_type, reference_number, title, description,
		       applicability, affected_aircraft_types, effective_date, compliance_deadline,
		       recurrence_interval, superseded_by, source_url, created_at, updated_at
		FROM compliance_directives
		WHERE 1=1`
	if len(clauses) > 0 {
		query += " AND " + strings.Join(clauses, " AND ")
	}
	args = append(args, limit, offset)
	query += " ORDER BY effective_date DESC LIMIT $" + itoa(len(args)-1) + " OFFSET $" + itoa(len(args))

	rows, err := r.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.ComplianceDirective
	for rows.Next() {
		d, err := scanDirective(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, d)
	}
	return items, rows.Err()
}

func (r *DirectiveRepository) CreateDirective(ctx context.Context, d domain.ComplianceDirective) (domain.ComplianceDirective, error) {
	row := r.DB.QueryRow(ctx, `
		INSERT INTO compliance_directives
			(id, org_id, authority_id, directive_type, reference_number, title, description,
			 applicability, affected_aircraft_types, effective_date, compliance_deadline,
			 recurrence_interval, superseded_by, source_url, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)
		RETURNING id, org_id, authority_id, directive_type, reference_number, title, description,
		          applicability, affected_aircraft_types, effective_date, compliance_deadline,
		          recurrence_interval, superseded_by, source_url, created_at, updated_at
	`, d.ID, d.OrgID, d.AuthorityID, d.DirectiveType, d.ReferenceNumber, d.Title, d.Description,
		d.Applicability, d.AffectedAircraftTypes, d.EffectiveDate, d.ComplianceDeadline,
		d.RecurrenceInterval, d.SupersededBy, d.SourceURL, d.CreatedAt, d.UpdatedAt)
	created, err := scanDirective(row)
	if err != nil {
		return domain.ComplianceDirective{}, TranslateError(err)
	}
	return created, nil
}

func (r *DirectiveRepository) UpdateDirective(ctx context.Context, d domain.ComplianceDirective) (domain.ComplianceDirective, error) {
	row := r.DB.QueryRow(ctx, `
		UPDATE compliance_directives
		SET title=$1, description=$2, applicability=$3, affected_aircraft_types=$4,
		    compliance_deadline=$5, recurrence_interval=$6, superseded_by=$7, source_url=$8, updated_at=$9
		WHERE id=$10
		RETURNING id, org_id, authority_id, directive_type, reference_number, title, description,
		          applicability, affected_aircraft_types, effective_date, compliance_deadline,
		          recurrence_interval, superseded_by, source_url, created_at, updated_at
	`, d.Title, d.Description, d.Applicability, d.AffectedAircraftTypes,
		d.ComplianceDeadline, d.RecurrenceInterval, d.SupersededBy, d.SourceURL, d.UpdatedAt, d.ID)
	updated, err := scanDirective(row)
	if err != nil {
		return domain.ComplianceDirective{}, TranslateError(err)
	}
	return updated, nil
}

func scanDirective(row pgx.Row) (domain.ComplianceDirective, error) {
	var d domain.ComplianceDirective
	if err := row.Scan(&d.ID, &d.OrgID, &d.AuthorityID, &d.DirectiveType, &d.ReferenceNumber,
		&d.Title, &d.Description, &d.Applicability, &d.AffectedAircraftTypes,
		&d.EffectiveDate, &d.ComplianceDeadline, &d.RecurrenceInterval,
		&d.SupersededBy, &d.SourceURL, &d.CreatedAt, &d.UpdatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return domain.ComplianceDirective{}, domain.ErrNotFound
		}
		return domain.ComplianceDirective{}, err
	}
	return d, nil
}

// --- Aircraft Directive Compliance ---

func (r *DirectiveRepository) GetAircraftCompliance(ctx context.Context, orgID, aircraftID, directiveID uuid.UUID) (domain.AircraftDirectiveCompliance, error) {
	row := r.DB.QueryRow(ctx, `
		SELECT id, org_id, aircraft_id, directive_id, status, compliance_date, next_due_date,
		       task_id, signed_off_by, signed_off_at, notes, created_at, updated_at
		FROM aircraft_directive_compliance
		WHERE org_id=$1 AND aircraft_id=$2 AND directive_id=$3
	`, orgID, aircraftID, directiveID)
	return scanAircraftCompliance(row)
}

func (r *DirectiveRepository) ListAircraftCompliance(ctx context.Context, filter ports.AircraftComplianceFilter) ([]domain.AircraftDirectiveCompliance, error) {
	clauses := make([]string, 0, 4)
	args := make([]any, 0, 4)
	add := func(condition string, value any) {
		args = append(args, value)
		clauses = append(clauses, condition+"$"+itoa(len(args)))
	}
	if filter.OrgID != nil {
		add("org_id=", *filter.OrgID)
	}
	if filter.AircraftID != nil {
		add("aircraft_id=", *filter.AircraftID)
	}
	if filter.Status != nil {
		add("status=", *filter.Status)
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	query := `
		SELECT id, org_id, aircraft_id, directive_id, status, compliance_date, next_due_date,
		       task_id, signed_off_by, signed_off_at, notes, created_at, updated_at
		FROM aircraft_directive_compliance
		WHERE 1=1`
	if len(clauses) > 0 {
		query += " AND " + strings.Join(clauses, " AND ")
	}
	args = append(args, limit, offset)
	query += " ORDER BY created_at DESC LIMIT $" + itoa(len(args)-1) + " OFFSET $" + itoa(len(args))

	rows, err := r.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.AircraftDirectiveCompliance
	for rows.Next() {
		c, err := scanAircraftCompliance(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, c)
	}
	return items, rows.Err()
}

func (r *DirectiveRepository) UpsertAircraftCompliance(ctx context.Context, c domain.AircraftDirectiveCompliance) (domain.AircraftDirectiveCompliance, error) {
	row := r.DB.QueryRow(ctx, `
		INSERT INTO aircraft_directive_compliance
			(id, org_id, aircraft_id, directive_id, status, compliance_date, next_due_date,
			 task_id, signed_off_by, signed_off_at, notes, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
		ON CONFLICT (org_id, aircraft_id, directive_id)
		DO UPDATE SET status=$5, compliance_date=$6, next_due_date=$7,
		              task_id=$8, signed_off_by=$9, signed_off_at=$10, notes=$11, updated_at=$13
		RETURNING id, org_id, aircraft_id, directive_id, status, compliance_date, next_due_date,
		          task_id, signed_off_by, signed_off_at, notes, created_at, updated_at
	`, c.ID, c.OrgID, c.AircraftID, c.DirectiveID, c.Status, c.ComplianceDate, c.NextDueDate,
		c.TaskID, c.SignedOffBy, c.SignedOffAt, c.Notes, c.CreatedAt, c.UpdatedAt)
	result, err := scanAircraftCompliance(row)
	if err != nil {
		return domain.AircraftDirectiveCompliance{}, TranslateError(err)
	}
	return result, nil
}

func scanAircraftCompliance(row pgx.Row) (domain.AircraftDirectiveCompliance, error) {
	var c domain.AircraftDirectiveCompliance
	if err := row.Scan(&c.ID, &c.OrgID, &c.AircraftID, &c.DirectiveID, &c.Status,
		&c.ComplianceDate, &c.NextDueDate, &c.TaskID, &c.SignedOffBy, &c.SignedOffAt,
		&c.Notes, &c.CreatedAt, &c.UpdatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return domain.AircraftDirectiveCompliance{}, domain.ErrNotFound
		}
		return domain.AircraftDirectiveCompliance{}, err
	}
	return c, nil
}

// --- Templates ---

func (r *DirectiveRepository) ListTemplates(ctx context.Context, authorityID uuid.UUID) ([]domain.ComplianceTemplate, error) {
	rows, err := r.DB.Query(ctx, `
		SELECT id, authority_id, template_code, name, description, required_fields, template_content, created_at
		FROM compliance_templates
		WHERE authority_id=$1
		ORDER BY template_code
	`, authorityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.ComplianceTemplate
	for rows.Next() {
		t, err := scanTemplate(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, t)
	}
	return items, rows.Err()
}

func (r *DirectiveRepository) GetTemplateByCode(ctx context.Context, authorityID uuid.UUID, code string) (domain.ComplianceTemplate, error) {
	row := r.DB.QueryRow(ctx, `
		SELECT id, authority_id, template_code, name, description, required_fields, template_content, created_at
		FROM compliance_templates
		WHERE authority_id=$1 AND template_code=$2
	`, authorityID, code)
	return scanTemplate(row)
}

func scanTemplate(row pgx.Row) (domain.ComplianceTemplate, error) {
	var t domain.ComplianceTemplate
	var fieldsJSON []byte
	if err := row.Scan(&t.ID, &t.AuthorityID, &t.TemplateCode, &t.Name, &t.Description,
		&fieldsJSON, &t.TemplateContent, &t.CreatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return domain.ComplianceTemplate{}, domain.ErrNotFound
		}
		return domain.ComplianceTemplate{}, err
	}
	if fieldsJSON != nil {
		_ = json.Unmarshal(fieldsJSON, &t.RequiredFields)
	}
	return t, nil
}
