-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
SET search_path = valuations_api, public;

ALTER TABLE valuations_api.valuations
    ADD COLUMN token_id numeric(78, 0);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
SET search_path = valuations_api, public;

ALTER TABLE valuations_api.valuations
    DROP COLUMN token_id;

-- +goose StatementEnd
