// Package repository contains all the functionality for working with the DB.
package repository

import (
	"database/sql"
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
	SystemWalletName        = "system"
)

func NewRepo(logger *zap.SugaredLogger, host string, port int, user, password, dbname string) (*Repo, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("connect database: %w", err)
	}

	err = MigrateUp(logger, db, "file://migrations")
	if err != nil {
		return nil, err
	}

	return NewRepoWithDB(logger, db)
}

func MigrateUp(logger *zap.SugaredLogger, db *sqlx.DB, path string) error {
	driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
	m, err := migrate.NewWithDatabaseInstance(path, "postgres", driver)
	if err != nil {
		return fmt.Errorf("migrate NewWithDatabaseInstance: %w", err)
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migrate Up: %w", err)
	}

	version, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migrate version: %w", err)
	}
	logger.With("version", version, "dirty", dirty).Info("Migrations applied")

	return nil
}

func NewRepoWithDB(logger *zap.SugaredLogger, db *sqlx.DB) (*Repo, error) {
	return &Repo{
		log: logger,
		db:  db,
	}, nil
}

func (r *Repo) CreateWallet(walletName string) error {
	r.log.With("wallet", walletName).Debug("CreateWallet")
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

func (r *Repo) GetWallet(walletName string) (*dto.Wallet, error) {
	r.log.With("wallet", walletName).Debug("GetWallet")
	const query = `
SELECT * 
FROM wallets 
WHERE name = $1 
`

	var dbWallet Wallet
	err := r.db.Get(&dbWallet, query, walletName)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("select: %w", err)
	}

	return &dto.Wallet{
		Name:    dbWallet.Name,
		Balance: dbWallet.Balance,
	}, nil
}

func (r *Repo) IncreaseWalletBalance(walletName string, amount uint64) error {
	r.log.With("wallet_name", walletName, "amount", amount).Debug("IncreaseWalletBalance")

	return r.RunWithTransaction(func(tx *sqlx.Tx) error {
		err := r.increaseWalletBalanceTx(tx, walletName, amount)
		if err != nil {
			return err
		}

		return r.insertOperation(tx, walletName, OperationTypeDeposit, amount, SystemWalletName)
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
		wallets[i].Name = dbWallets[i].Name
		wallets[i].Balance = dbWallets[i].Balance
	}

	return wallets, nil
}

func (r *Repo) TransferTx(tx *sqlx.Tx, walletFrom, walletTo string, amount uint64) error {
	r.log.With("wallet_from", walletFrom, "wallet_to", walletTo, "amount", amount).Debug("TransferTx")

	err := r.decreaseWalletBalanceTx(tx, walletFrom, amount)
	if err != nil {
		return fmt.Errorf("decrease wallet balance: %w", err)
	}

	err = r.insertOperation(tx, walletFrom, OperationTypeWithdrawal, amount, walletTo)
	if err != nil {
		return fmt.Errorf("insert operation: %w", err)
	}

	err = r.increaseWalletBalanceTx(tx, walletTo, amount)
	if err != nil {
		return fmt.Errorf("increase wallet balance: %w", err)
	}

	err = r.insertOperation(tx, walletTo, OperationTypeDeposit, amount, walletFrom)
	if err != nil {
		return fmt.Errorf("insert operation: %w", err)
	}

	return nil
}

func (r *Repo) decreaseWalletBalanceTx(tx *sqlx.Tx, walletName string, amount uint64) error {
	r.log.With("wallet", walletName, "amount", amount).Debug("decreaseWalletBalanceTx")
	const query = `
UPDATE wallets
SET balance = balance - $2
WHERE name = $1 AND balance >= $2
`

	_, err := tx.Exec(query, walletName, amount)
	if err != nil {
		return fmt.Errorf("update wallets: %w", err)
	}

	return nil
}

func (r *Repo) increaseWalletBalanceTx(tx *sqlx.Tx, walletName string, amount uint64) error {
	r.log.With("wallet", walletName, "amount", amount).Debug("increaseWalletBalanceTx")
	const query = `
UPDATE wallets
SET balance = balance + $2
WHERE name = $1
`

	_, err := tx.Exec(query, walletName, amount)
	if err != nil {
		return fmt.Errorf("update wallets: %w", err)
	}

	return nil
}

func (r *Repo) insertOperation(tx *sqlx.Tx, wallet, opType string, amount uint64, otherWallet string) error {
	r.log.With("wallet", wallet, "type", opType, "amount", amount, "other", otherWallet).Debug("insertOperation")

	const query = `
INSERT INTO operations (wallet, type, amount, other_wallet)
VALUES ($1, $2, $3, $4)
`

	_, err := tx.Exec(query, wallet, opType, amount, otherWallet)
	if err != nil {
		return fmt.Errorf("update wallets: %w", err)
	}

	return nil
}

func (r *Repo) GetOperations(walletName string) ([]dto.Operation, error) {
	r.log.With("wallet", walletName).Debug("GetOperations")

	const query = `
SELECT * 
FROM operations 
WHERE wallet = $1
`

	dbOperations := make([]Operation, 0)
	err := r.db.Select(&dbOperations, query, walletName)
	if err != nil {
		return nil, fmt.Errorf("select for update: %w", err)
	}

	operations := make([]dto.Operation, len(dbOperations))
	for i := range dbOperations {
		operations[i].Wallet = dbOperations[i].Wallet
		operations[i].Type = dbOperations[i].Type
		operations[i].Amount.SetAmount(dbOperations[i].Amount)
		operations[i].OtherWallet = dbOperations[i].OtherWallet
	}

	return operations, nil
}
