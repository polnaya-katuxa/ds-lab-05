-- +goose Up
-- +goose StatementBegin
CREATE TABLE rental
(
    id          SERIAL PRIMARY KEY,
    rental_uid  uuid UNIQUE              NOT NULL,
    username    VARCHAR(80)              NOT NULL,
    payment_uid uuid                     NOT NULL,
    car_uid     uuid                     NOT NULL,
    date_from   TIMESTAMP WITH TIME ZONE NOT NULL,
    date_to     TIMESTAMP WITH TIME ZONE NOT NULL,
    status      VARCHAR(20)              NOT NULL
        CHECK (status IN ('IN_PROGRESS', 'FINISHED', 'CANCELED'))
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE rental IF EXISTS;
-- +goose StatementEnd
