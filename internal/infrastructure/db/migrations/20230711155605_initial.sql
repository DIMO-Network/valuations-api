-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';

SET search_path = valuations_api, public;

CREATE TABLE IF NOT EXISTS valuations_api.valuations
(
    id character(27) COLLATE pg_catalog."default" NOT NULL,
    device_definition_id character(27) COLLATE pg_catalog."default",
    vin text COLLATE pg_catalog."default" NOT NULL,
    user_device_id character(27) COLLATE pg_catalog."default",
    vin_metadata json,
    offer_metadata json,
    autocheck_metadata json,
    build_metadata json,
    cargurus_metadata json,
    carvana_metadata json,
    carmax_metadata json,
    carstory_metadata json,
    edmunds_metadata json,
    tmv_metadata json,
    kbb_metadata json,
    vroom_metadata json,
    created_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
                             pricing_metadata jsonb,
                             blackbook_metadata jsonb,
                             request_metadata jsonb,
                             vincario_metadata jsonb,
                             CONSTRAINT drivly_data_pkey PRIMARY KEY (id)
    );

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';

SET search_path = valuations_api, public;
DROP TABLE valuations_api.valuations

-- +goose StatementEnd
