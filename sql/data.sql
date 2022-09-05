CREATE TABLE "wallet" (
                          "id" bigserial PRIMARY KEY,
                          "owner" varchar NOT NULL,
                          "balance" bigint NOT NULL,
                          "created_at" timestamptz NOT NULL DEFAULT 'now()'

);
CREATE TABLE "transaction"(
                              "id" bigserial PRIMARY KEY,
                              "type" varchar NOT NULL,
                              "sum" bigint NOT NULL,
                              "created_at" timestamptz NOT NULL DEFAULT 'now()',
                              "from_account_id" int NOT NULL,
                              "to_account_id" int NOT NULL

);
