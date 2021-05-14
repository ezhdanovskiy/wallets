CREATE TYPE operation_type AS ENUM ('deposit', 'withdraw');

CREATE TABLE "history"
(
    "id"         bigserial PRIMARY KEY,
    "wallet"     varchar        NOT NULL,
    "operation"  operation_type NOT NULL,
    "amount"     bigint         NOT NULL DEFAULT (0),
    "created_at" timestamptz    NOT NULL DEFAULT (now())
);

/*
INSERT INTO history (wallet, operation, amount)
VALUES ('wallet01', 100000),
       ('wallet02', 100000);
 */