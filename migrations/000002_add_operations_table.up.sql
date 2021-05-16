CREATE TYPE operation_type AS ENUM ('deposit', 'withdrawal');

CREATE TABLE "operations"
(
    "id"           bigserial PRIMARY KEY,
    "wallet"       varchar        NOT NULL DEFAULT 'system',
    "type"         operation_type NOT NULL,
    "amount"       bigint         NOT NULL DEFAULT 0,
    "other_wallet" varchar        NOT NULL DEFAULT 'system',
    "created_at"   timestamptz    NOT NULL DEFAULT now()
);


INSERT INTO operations (wallet, type, amount)
VALUES ('wallet01', 'deposit', 100000),
       ('wallet02', 'deposit', 100000);
