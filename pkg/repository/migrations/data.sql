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
    id              bigserial PRIMARY KEY,
    type            varchar        NOT NULL,
    sum             numeric(10, 2) NOT NULL,
    created_at      timestamptz    NOT NULL DEFAULT now(),
    from_account_id int            NOT NULL,
    to_account_id   int            NOT NULL

);
