-- +goose Up
-- +goose StatementBegin
CREATE INDEX idx_users_deleted_at ON users (deleted_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_users_deleted_at;
-- +goose StatementEnd
