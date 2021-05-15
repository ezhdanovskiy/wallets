// Package repository contains all the functionality for working with the DB.
package repository

import (
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"go.uber.org/zap"

	"github.com/ezhdanovskiy/wallets/internal/dto"
)

type Repo struct {
	log *zap.SugaredLogger
	db  *sqlx.DB
}

const (
	OperationTypeDeposit    = "deposit"
	OperationTypeWithdrawal = "withdrawal"
)

func NewRepo(logger *zap.SugaredLogger, host string, port int, user, password, dbname string) (*Repo, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("connect database: %w", err)
	}

	driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
	m, err := migrate.NewWithDatabaseInstance("file://migrations", "postgres", driver)
	if err != nil {
		return nil, fmt.Errorf("migrate NewWithDatabaseInstance: %w", err)
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return nil, fmt.Errorf("migrate Up: %w", err)
	}

	version, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNoChange {
		return nil, fmt.Errorf("migrate version: %w", err)
	}
	logger.With("version", version, "dirty", dirty).Info("Migrations applied")

	return NewRepoWithDB(logger, db)
}

func NewRepoWithDB(logger *zap.SugaredLogger, db *sqlx.DB) (*Repo, error) {
	return &Repo{
		log: logger,
		db:  db,
	}, nil
}

func (r *Repo) CreateWallet(walletName string) error {
	r.log.With("wallet_name", walletName).Debug("CreateWallet")
	const query = `
INSERT INTO wallets (name) 
VALUES ($1) 
ON CONFLICT DO NOTHING
`

	_, err := r.db.Exec(query, walletName)
	if err != nil {
		return fmt.Errorf("insert wallets: %w", err)
	}

	return nil
}

func (r *Repo) IncreaseWalletBalance(walletName string, amount uint64) error {
	r.log.With("wallet_name", walletName, "amount", amount).Debug("IncreaseWalletAmount")

	return r.RunWithTransaction(func(tx *sqlx.Tx) error {
		err := r.IncreaseWalletBalanceTx(tx, walletName, amount)
		if err != nil {
			return err
		}

		return nil
	})
}

func (r *Repo) RunWithTransaction(f func(tx *sqlx.Tx) error) error {
	r.log.Debug("RunWithTransaction")

	tx, err := r.db.Beginx()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	fErr := f(tx)
	if fErr != nil {
		if err = tx.Rollback(); err != nil {
			r.log.Errorf("rollback: %s", err)
		}
		return fErr
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}

func (r *Repo) GetWalletsForUpdateTx(tx *sqlx.Tx, walletNames []string) ([]dto.Wallet, error) {
	r.log.With("wallets", walletNames).Debug("GetWalletsForUpdateTx")

	const querySrc = `
SELECT * 
FROM wallets 
WHERE name IN (?)
FOR UPDATE 
`

	query, args, err := sqlx.In(querySrc, walletNames)
	if err != nil {
		return nil, fmt.Errorf("select for update: %w", err)
	}

	dbWallets := make([]Wallet, 0)
	err = tx.Select(&dbWallets, r.db.Rebind(query), args...)
	if err != nil {
		return nil, fmt.Errorf("select for update: %w", err)
	}

	wallets := make([]dto.Wallet, len(dbWallets))
	for i := range dbWallets {
		wallets[i].ID = dbWallets[i].ID
		wallets[i].Name = dbWallets[i].Name
		wallets[i].Balance = dbWallets[i].Balance
	}

	return wallets, nil
}

func (r *Repo) DecreaseWalletBalanceTx(tx *sqlx.Tx, walletName string, amount uint64) error {
	r.log.With("wallet_name", walletName, "amount", amount).Debug("DecreaseWalletBalanceTx")
	const query = `
UPDATE wallets
SET balance = balance - $2
WHERE name = $1 AND balance >= $2
`

	_, err := tx.Exec(query, walletName, amount)
	if err != nil {
		return fmt.Errorf("update wallets: %w", err)
	}

	return r.insertOperation(tx, walletName, OperationTypeWithdrawal, amount)
}

func (r *Repo) IncreaseWalletBalanceTx(tx *sqlx.Tx, walletName string, amount uint64) error {
	r.log.With("wallet_name", walletName, "amount", amount).Debug("IncreaseWalletBalanceTx")
	const query = `
UPDATE wallets
SET balance = balance + $2
WHERE name = $1
`

	_, err := tx.Exec(query, walletName, amount)
	if err != nil {
		return fmt.Errorf("update wallets: %w", err)
	}

	return r.insertOperation(tx, walletName, OperationTypeDeposit, amount)
}

func (r *Repo) insertOperation(tx *sqlx.Tx, walletName, opType string, amount uint64) error {
	r.log.With("wallet_name", walletName, "amount", amount).Debug("insertOperation")

	const query = `
INSERT INTO operations (wallet_to, operation, amount)
VALUES ($1, $2, $3)
`

	_, err := tx.Exec(query, walletName, opType, amount)
	if err != nil {
		return fmt.Errorf("update wallets: %w", err)
	}

	return nil
}
