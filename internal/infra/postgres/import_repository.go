package postgres

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/aeromaintain/amss/internal/app/ports"
	"github.com/aeromaintain/amss/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ImportRepository struct {
	DB *pgxpool.Pool
}

func (r *ImportRepository) GetByID(ctx context.Context, orgID, id uuid.UUID) (domain.Import, error) {
	if r == nil || r.DB == nil {
		return domain.Import{}, domain.ErrNotFound
	}
	var row pgx.Row
	if orgID == uuid.Nil {
		row = r.DB.QueryRow(ctx, `
			SELECT id, org_id, type, status, file_name, file_path, created_by, summary, created_at, updated_at
			FROM imports
			WHERE id=$1
		`, id)
	} else {
		row = r.DB.QueryRow(ctx, `
			SELECT id, org_id, type, status, file_name, file_path, created_by, summary, created_at, updated_at
			FROM imports
			WHERE org_id=$1 AND id=$2
		`, orgID, id)
	}
	return scanImport(row)
}

func (r *ImportRepository) Create(ctx context.Context, imp domain.Import) (domain.Import, error) {
	if r == nil || r.DB == nil {
		return domain.Import{}, domain.ErrNotFound
	}
	row := r.DB.QueryRow(ctx, `
		INSERT INTO imports
			(id, org_id, type, status, file_name, file_path, created_by, summary, created_at, updated_at)
		VALUES
			($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		RETURNING id, org_id, type, status, file_name, file_path, created_by, summary, created_at, updated_at
	`, imp.ID, imp.OrgID, imp.Type, imp.Status, imp.FileName, imp.FilePath, imp.CreatedBy, encodeJSON(imp.Summary), imp.CreatedAt, imp.UpdatedAt)
	created, err := scanImport(row)
	if err != nil {
		return domain.Import{}, TranslateError(err)
	}
	return created, nil
}

func (r *ImportRepository) Update(ctx context.Context, imp domain.Import) (domain.Import, error) {
	if r == nil || r.DB == nil {
		return domain.Import{}, domain.ErrNotFound
	}
	row := r.DB.QueryRow(ctx, `
		UPDATE imports
		SET status=$1, file_name=$2, file_path=$3, summary=$4, updated_at=$5
		WHERE org_id=$6 AND id=$7
		RETURNING id, org_id, type, status, file_name, file_path, created_by, summary, created_at, updated_at
	`, imp.Status, imp.FileName, imp.FilePath, encodeJSON(imp.Summary), imp.UpdatedAt, imp.OrgID, imp.ID)
	updated, err := scanImport(row)
	if err != nil {
		return domain.Import{}, TranslateError(err)
	}
	return updated, nil
}

func (r *ImportRepository) UpdateStatus(ctx context.Context, orgID, id uuid.UUID, status domain.ImportStatus, summary map[string]any, updatedAt time.Time) error {
	if r == nil || r.DB == nil {
		return domain.ErrNotFound
	}
	_, err := r.DB.Exec(ctx, `
		UPDATE imports
		SET status=$1, summary=$2, updated_at=$3
		WHERE org_id=$4 AND id=$5
	`, status, encodeJSON(summary), updatedAt, orgID, id)
	return TranslateError(err)
}

type ImportRowRepository struct {
	DB *pgxpool.Pool
}

func (r *ImportRowRepository) Create(ctx context.Context, row domain.ImportRow) error {
	if r == nil || r.DB == nil {
		return domain.ErrNotFound
	}
	_, err := r.DB.Exec(ctx, `
		INSERT INTO import_rows
			(id, org_id, import_id, row_number, raw, status, errors, created_at, updated_at)
		VALUES
			($1,$2,$3,$4,$5,$6,$7,$8,$9)
	`, row.ID, row.OrgID, row.ImportID, row.RowNumber, encodeJSON(row.Raw), row.Status, encodeJSON(row.Errors), row.CreatedAt, row.UpdatedAt)
	return TranslateError(err)
}

func (r *ImportRowRepository) Update(ctx context.Context, row domain.ImportRow) error {
	if r == nil || r.DB == nil {
		return domain.ErrNotFound
	}
	_, err := r.DB.Exec(ctx, `
		UPDATE import_rows
		SET status=$1, errors=$2, updated_at=$3
		WHERE org_id=$4 AND import_id=$5 AND id=$6
	`, row.Status, encodeJSON(row.Errors), row.UpdatedAt, row.OrgID, row.ImportID, row.ID)
	return TranslateError(err)
}

func (r *ImportRowRepository) ListByImport(ctx context.Context, filter ports.ImportRowFilter) ([]domain.ImportRow, error) {
	if r == nil || r.DB == nil {
		return nil, nil
	}
	clauses := make([]string, 0, 2)
	args := make([]any, 0, 6)
	add := func(condition string, value any) {
		args = append(args, value)
		clauses = append(clauses, condition+"$"+itoa(len(args)))
	}
	add("org_id=", filter.OrgID)
	add("import_id=", filter.ImportID)
	if filter.Status != nil {
		add("status=", *filter.Status)
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 100
	}
	if limit > 500 {
		limit = 500
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	query := `
		SELECT id, org_id, import_id, row_number, raw, status, errors, created_at, updated_at
		FROM import_rows
		WHERE ` + strings.Join(clauses, " AND ")
	args = append(args, limit, offset)
	query += " ORDER BY row_number ASC LIMIT $" + itoa(len(args)-1) + " OFFSET $" + itoa(len(args))

	rows, err := r.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.ImportRow
	for rows.Next() {
		item, err := scanImportRow(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func scanImport(row pgx.Row) (domain.Import, error) {
	var imp domain.Import
	var summary []byte
	if err := row.Scan(&imp.ID, &imp.OrgID, &imp.Type, &imp.Status, &imp.FileName, &imp.FilePath, &imp.CreatedBy, &summary, &imp.CreatedAt, &imp.UpdatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return domain.Import{}, domain.ErrNotFound
		}
		return domain.Import{}, err
	}
	if len(summary) > 0 {
		var payload map[string]any
		if err := json.Unmarshal(summary, &payload); err == nil {
			imp.Summary = payload
		}
	}
	return imp, nil
}

func scanImportRow(row pgx.Row) (domain.ImportRow, error) {
	var item domain.ImportRow
	var raw []byte
	var errs []byte
	if err := row.Scan(&item.ID, &item.OrgID, &item.ImportID, &item.RowNumber, &raw, &item.Status, &errs, &item.CreatedAt, &item.UpdatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return domain.ImportRow{}, domain.ErrNotFound
		}
		return domain.ImportRow{}, err
	}
	if len(raw) > 0 {
		var payload map[string]any
		if err := json.Unmarshal(raw, &payload); err == nil {
			item.Raw = payload
		}
	}
	if len(errs) > 0 {
		var errorsList []string
		if err := json.Unmarshal(errs, &errorsList); err == nil {
			item.Errors = errorsList
		}
	}
	return item, nil
}

func encodeJSON(value any) []byte {
	if value == nil {
		return nil
	}
	data, err := json.Marshal(value)
	if err != nil {
		return nil
	}
	return data
}
