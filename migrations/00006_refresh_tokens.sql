-- +goose Up
CREATE TABLE IF NOT EXISTS refresh_tokens (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES organizations(id),
  user_id uuid NOT NULL,
  token_hash text NOT NULL,
  token_id text NOT NULL,
  expires_at timestamptz NOT NULL,
  revoked_at timestamptz,
  rotated_from uuid,
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (org_id, token_hash),
  UNIQUE (org_id, token_id),
  FOREIGN KEY (org_id, user_id) REFERENCES users(org_id, id),
  FOREIGN KEY (rotated_from) REFERENCES refresh_tokens(id)
);

CREATE INDEX IF NOT EXISTS refresh_tokens_user_idx ON refresh_tokens (org_id, user_id, revoked_at);
CREATE INDEX IF NOT EXISTS refresh_tokens_expires_idx ON refresh_tokens (expires_at);

-- +goose Down
DROP INDEX IF EXISTS refresh_tokens_expires_idx;
DROP INDEX IF EXISTS refresh_tokens_user_idx;
DROP TABLE IF EXISTS refresh_tokens;
