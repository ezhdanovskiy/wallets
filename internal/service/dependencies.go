package service

import (
	"github.com/jmoiron/sqlx"

	"github.com/ezhdanovskiy/wallets/internal/dto"
)

type Repository interface {
	CreateWallet(walletName string) error
	GetWallet(walletName string) (*dto.Wallet, error)
	IncreaseWalletBalance(walletName string, amount uint64) error
	GetOperations(dto.OperationsFilter) ([]dto.Operation, error)

	RunWithTransaction(f func(tx *sqlx.Tx) error) error
	GetWalletsForUpdateTx(tx *sqlx.Tx, walletNames []string) ([]dto.Wallet, error)
	TransferTx(tx *sqlx.Tx, walletFrom, walletTo string, amount uint64) error
}

//go:generate mockgen -destination=./mocks/repository_mock.go -package=mocks . Repository
