CREATE TABLE "wallets"
(
    "name"       varchar     PRIMARY KEY,
    "balance"    bigint      NOT NULL DEFAULT (0),
    "created_at" timestamptz NOT NULL DEFAULT (now()),
    "updated_at" timestamptz NOT NULL DEFAULT (now())
);


INSERT INTO wallets (name, balance)
VALUES ('wallet01', 100000),
       ('wallet02', 100000);
