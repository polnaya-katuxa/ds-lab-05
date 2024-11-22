-- +goose Up
-- +goose StatementBegin
CREATE TABLE payment
(
    id          SERIAL PRIMARY KEY,
    payment_uid uuid        NOT NULL,
    status      VARCHAR(20) NOT NULL
        CHECK (status IN ('PAID', 'CANCELED')),
    price       INT         NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE payment IF EXISTS;
-- +goose StatementEnd
