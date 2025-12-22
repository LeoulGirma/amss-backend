-- +goose Up
ALTER TABLE imports ADD COLUMN file_path text NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE imports DROP COLUMN IF EXISTS file_path;
