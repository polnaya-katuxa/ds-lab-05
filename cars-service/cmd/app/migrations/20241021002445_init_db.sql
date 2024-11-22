-- +goose Up
-- +goose StatementBegin
CREATE TABLE cars
(
    id                  SERIAL PRIMARY KEY,
    car_uid             uuid UNIQUE NOT NULL,
    brand               VARCHAR(80) NOT NULL,
    model               VARCHAR(80) NOT NULL,
    registration_number VARCHAR(20) NOT NULL,
    power               INT,
    price               INT         NOT NULL,
    type                VARCHAR(20)
        CHECK (type IN ('SEDAN', 'SUV', 'MINIVAN', 'ROADSTER')),
    availability        BOOLEAN     NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE cars IF EXISTS;
-- +goose StatementEnd
