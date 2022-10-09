-- noinspection SqlNoDataSourceInspectionForFile

-- +migrate Up
CREATE TABLE IF NOT EXISTS wallet
(
    id         bigserial PRIMARY KEY,
    owner      varchar        NOT NULL,
    balance    numeric(10, 2) NOT NULL,
    created_at timestamptz    NOT NULL DEFAULT now(),
    updated_at timestamptz    NOT NULL DEFAULT now()

);
CREATE TABLE IF NOT EXISTS transaction
(
    id        bigserial PRIMARY KEY,
    uuid      text UNIQUE NOT NULL,
    from_id   integer     NOT NULL,
    to_id     integer    DEFAULT NULL,
    operation varchar NOT NULL,
    sum       numeric(10, 2) NOT NULL,
    date      timestamptz NOT NULL DEFAULT now()
);
-- +migrate Down
DROP TABLE wallet CASCADE;
DROP TABLE transaction CASCADE;
