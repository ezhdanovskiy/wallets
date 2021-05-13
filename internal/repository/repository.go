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

func (r *Repo) CreateWallet(wallet dto.Wallet) error {
	r.log.With("wallet", wallet).Debug("CreateWallet")
	const query = `
INSERT INTO wallets (name) 
VALUES ($1) 
ON CONFLICT DO NOTHING
`

	_, err := r.db.Exec(query, wallet.Name)
	if err != nil {
		return fmt.Errorf("insert: %w", err)
	}

	return nil
}
