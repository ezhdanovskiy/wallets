CREATE TYPE operation_type AS ENUM ('deposit', 'withdrawal');

CREATE TABLE "operations"
(
    "id"          bigserial PRIMARY KEY,
    "wallet_from" varchar        NOT NULL DEFAULT 'system',
    "wallet_to"   varchar        NOT NULL DEFAULT 'system',
    "operation"   operation_type NOT NULL,
    "amount"      bigint         NOT NULL DEFAULT 0,
    "created_at"  timestamptz    NOT NULL DEFAULT now()
);


INSERT INTO operations (wallet_from, wallet_to, operation, amount)
VALUES ('system', 'wallet01', 'deposit', 100000),
       ('system', 'wallet02', 'deposit', 100000);
