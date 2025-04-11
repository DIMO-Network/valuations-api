-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
alter table valuations add column definition_id text;
alter table valuations drop column user_device_id;

create table valuations_api.geodecoded_location
(
    token_id    bigint                              not null
        constraint geodecoded_location_pk
            primary key,
    postal_code text,
    country text, 
    created_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
alter table valuations drop column definition_id;
alter table valuations add column user_device_id char(27);
-- +goose StatementEnd
