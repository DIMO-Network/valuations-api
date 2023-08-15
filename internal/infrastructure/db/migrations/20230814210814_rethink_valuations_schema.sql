-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
SET search_path = valuations_api, public;

alter table valuations
    drop column vin_metadata;

alter table valuations
    drop column autocheck_metadata;

alter table valuations
    drop column build_metadata;

alter table valuations
    drop column cargurus_metadata;

alter table valuations
    drop column carvana_metadata;

alter table valuations
    drop column carmax_metadata;

alter table valuations
    drop column carstory_metadata;

alter table valuations
    drop column tmv_metadata;

alter table valuations
    drop column kbb_metadata;

alter table valuations
    drop column vroom_metadata;

alter table valuations
    rename column pricing_metadata to drivly_pricing_metadata;

alter table valuations
    drop column blackbook_metadata;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
SET search_path = valuations_api, public;

-- +goose StatementEnd
