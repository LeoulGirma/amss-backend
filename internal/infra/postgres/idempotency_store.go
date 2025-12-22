package postgres

import (
	"context"
	"time"

	"github.com/aeromaintain/amss/internal/app/ports"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type IdempotencyStore struct {
	DB *pgxpool.Pool
}

func (s *IdempotencyStore) Get(ctx context.Context, orgID uuid.UUID, key, endpoint string) (ports.IdempotencyRecord, bool, error) {
	if s == nil || s.DB == nil {
		return ports.IdempotencyRecord{}, false, nil
	}
	row := s.DB.QueryRow(ctx, `
		SELECT request_hash, response_body, status_code, created_at, expires_at
		FROM idempotency_keys
		WHERE org_id=$1 AND key=$2 AND endpoint=$3
	`, orgID, key, endpoint)

	var record ports.IdempotencyRecord
	var body []byte
	if err := row.Scan(&record.RequestHash, &body, &record.StatusCode, &record.CreatedAt, &record.ExpiresAt); err != nil {
		if err == pgx.ErrNoRows {
			return ports.IdempotencyRecord{}, false, nil
		}
		return ports.IdempotencyRecord{}, false, err
	}
	if len(body) > 0 {
		record.Response = body
	}
	record.OrgID = orgID
	record.Key = key
	record.Endpoint = endpoint
	return record, true, nil
}

func (s *IdempotencyStore) CreatePlaceholder(ctx context.Context, record ports.IdempotencyRecord) error {
	if s == nil || s.DB == nil {
		return nil
	}
	_, err := s.DB.Exec(ctx, `
		INSERT INTO idempotency_keys
			(org_id, key, endpoint, request_hash, response_body, status_code, created_at, expires_at)
		VALUES
			($1, $2, $3, $4, NULL, 0, $5, $6)
	`, record.OrgID, record.Key, record.Endpoint, record.RequestHash, record.CreatedAt, record.ExpiresAt)
	return err
}

func (s *IdempotencyStore) UpdateResponse(ctx context.Context, orgID uuid.UUID, key, endpoint string, statusCode int, response []byte) error {
	if s == nil || s.DB == nil {
		return nil
	}
	var jsonValue []byte
	if len(response) > 0 {
		jsonValue = response
	}
	_, err := s.DB.Exec(ctx, `
		UPDATE idempotency_keys
		SET response_body=$4, status_code=$5
		WHERE org_id=$1 AND key=$2 AND endpoint=$3
	`, orgID, key, endpoint, jsonValue, statusCode)
	return err
}

func (s *IdempotencyStore) CleanupExpired(ctx context.Context, before time.Time) error {
	if s == nil || s.DB == nil {
		return nil
	}
	_, err := s.DB.Exec(ctx, `DELETE FROM idempotency_keys WHERE expires_at < $1`, before)
	return err
}
