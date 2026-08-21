-- +goose Up
-- +goose StatementBegin
ALTER TABLE role_permissions
    ADD CONSTRAINT fk_role_permissions_role_id
    FOREIGN KEY (role_id)
    REFERENCES roles(id)
    ON DELETE CASCADE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE role_permissions
    DROP CONSTRAINT IF EXISTS fk_role_permissions_role_id;
-- +goose StatementEnd
