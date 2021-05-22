// Package repository contains all the functionality for working with the DB.
package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"go.uber.org/zap"

	"github.com/ezhdanovskiy/wallets/internal/config"
	"github.com/ezhdanovskiy/wallets/internal/consts"
	"github.com/ezhdanovskiy/wallets/internal/dto"
)

type Repo struct {
	log *zap.SugaredLogger
	db  *sqlx.DB
}

// NewRepo creates instance of repository using config and applies migrations.
func NewRepo(logger *zap.SugaredLogger, cfg config.DB) (*Repo, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName)

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("connect database: %w", err)
	}

	err = MigrateUp(logger, db, "file://"+cfg.MigrationsPath)
	if err != nil {
		return nil, err
	}

	return NewRepoWithDB(logger, db)
}

// MigrateUp applies migrations to DB.
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

// NewRepoWithDB creates instance of repository using existing DB.
func NewRepoWithDB(logger *zap.SugaredLogger, db *sqlx.DB) (*Repo, error) {
	return &Repo{
		log: logger,
		db:  db,
	}, nil
}

// CreateWallet creates new wallet with unique name,
// or do nothing if wallet already exists.
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

// GetWallet selects wallet by name.
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

// IncreaseWalletBalance runs two operations in transaction:
// - increases wallet balance;
// - add new operation with type deposit.
func (r *Repo) IncreaseWalletBalance(walletName string, amount uint64) error {
	r.log.With("wallet_name", walletName, "amount", amount).Debug("IncreaseWalletBalance")

	return r.RunWithTransaction(func(tx *sqlx.Tx) error {
		err := r.increaseWalletBalanceTx(tx, walletName, amount)
		if err != nil {
			return err
		}

		return r.insertOperation(tx, walletName, consts.OperationTypeDeposit, amount, consts.SystemWalletName)
	})
}

// RunWithTransaction runs the given function inside a transaction.
func (r *Repo) RunWithTransaction(f func(tx *sqlx.Tx) error) error {
	r.log.Debug("RunWithTransaction")

	tx, err := r.db.BeginTxx(context.Background(), &sql.TxOptions{Isolation: sql.LevelSerializable})
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

// GetWalletsForUpdateTx selects wallets and obtains a lock for them at the database level using transaction.
// It will wait if some of the required wallets already locked in another goroutine.
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

// TransferTx runs four operations using transaction:
// - decreases balance of wallet_from if there is enough money;
// - add new operation with type withdrawal for wallet_from;
// - increases balance of wallet_to;
// - add new operation with type deposit for wallet_to.
func (r *Repo) TransferTx(tx *sqlx.Tx, walletFrom, walletTo string, amount uint64) error {
	r.log.With("wallet_from", walletFrom, "wallet_to", walletTo, "amount", amount).Debug("TransferTx")

	err := r.decreaseWalletBalanceTx(tx, walletFrom, amount)
	if err != nil {
		return fmt.Errorf("decrease wallet balance: %w", err)
	}

	err = r.insertOperation(tx, walletFrom, consts.OperationTypeWithdrawal, amount, walletTo)
	if err != nil {
		return fmt.Errorf("insert operation: %w", err)
	}

	err = r.increaseWalletBalanceTx(tx, walletTo, amount)
	if err != nil {
		return fmt.Errorf("increase wallet balance: %w", err)
	}

	err = r.insertOperation(tx, walletTo, consts.OperationTypeDeposit, amount, walletFrom)
	if err != nil {
		return fmt.Errorf("insert operation: %w", err)
	}

	return nil
}

func (r *Repo) decreaseWalletBalanceTx(tx *sqlx.Tx, walletName string, amount uint64) error {
	r.log.With("wallet", walletName, "amount", amount).Debug("decreaseWalletBalanceTx")
	const query = `
UPDATE wallets
SET balance = balance - $2, updated_at = now()
WHERE name = $1 AND balance >= $2
`

	res, err := tx.Exec(query, walletName, amount)
	if err != nil {
		return fmt.Errorf("update wallets: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}

	if rowsAffected != 1 {
		return fmt.Errorf("%s balance can't be decreased on this amount", walletName)
	}

	return nil
}

func (r *Repo) increaseWalletBalanceTx(tx *sqlx.Tx, walletName string, amount uint64) error {
	r.log.With("wallet", walletName, "amount", amount).Debug("increaseWalletBalanceTx")
	const query = `
UPDATE wallets
SET balance = balance + $2, updated_at = now()
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

// GetOperations selects operations for specified wallet using filter.
// Operations ordered by time.
func (r *Repo) GetOperations(filter dto.OperationsFilter) ([]dto.Operation, error) {
	r.log.With("wallet", filter.Wallet).Debug("GetOperations")

	queryTempl := `
SELECT * 
FROM operations 
WHERE %s
ORDER BY created_at
`

	namedArgs := make(map[string]interface{}) // Prepare named parameters.
	var whereParts []string                   // Generate where clause.

	whereParts = append(whereParts, "wallet = :wallet")
	namedArgs["wallet"] = filter.Wallet

	if len(filter.Type) != 0 {
		whereParts = append(whereParts, "type = :type")
		namedArgs["type"] = filter.Type
	}

	if filter.StartDate > 0 {
		whereParts = append(whereParts, "EXTRACT(EPOCH FROM created_at) >= :start_date")
		namedArgs["start_date"] = filter.StartDate
	}

	if filter.EndDate > 0 {
		whereParts = append(whereParts, "EXTRACT(EPOCH FROM created_at) <= :end_date")
		namedArgs["end_date"] = filter.EndDate
	}

	if filter.Limit > 0 {
		queryTempl += "LIMIT :limit\n"
		namedArgs["limit"] = filter.Limit
	}

	if filter.Offset > 0 {
		queryTempl += "OFFSET :offset\n"
		namedArgs["offset"] = filter.Offset
	}

	where := strings.Join(whereParts, " AND ")
	query := fmt.Sprintf(queryTempl, where)

	query, args, err := sqlx.Named(query, namedArgs)
	if err != nil {
		return nil, fmt.Errorf("sqlx named: %w", err)
	}

	query = r.db.Rebind(query)

	r.log.With("query", query, "args", args).Debug("select operations")
	dbOperations := make([]Operation, 0)
	err = r.db.Select(&dbOperations, query, args...)
	if err != nil {
		return nil, fmt.Errorf("select for update: %w", err)
	}

	operations := make([]dto.Operation, len(dbOperations))
	for i := range dbOperations {
		operations[i].Wallet = dbOperations[i].Wallet
		operations[i].Type = dbOperations[i].Type
		operations[i].Amount.SetAmount(dbOperations[i].Amount)
		operations[i].OtherWallet = dbOperations[i].OtherWallet
		operations[i].Timestamp = dbOperations[i].CreatedAt
	}

	return operations, nil
}
