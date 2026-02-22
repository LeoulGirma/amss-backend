-- +goose Up
-- +goose StatementBegin
DO $$ BEGIN
  ALTER TABLE imports ADD COLUMN file_path text NOT NULL DEFAULT '';
EXCEPTION WHEN duplicate_column THEN NULL;
END $$;
-- +goose StatementEnd

-- +goose Down
ALTER TABLE imports DROP COLUMN IF EXISTS file_path;
