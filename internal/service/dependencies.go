package service

import (
	"github.com/jmoiron/sqlx"

	"github.com/ezhdanovskiy/wallets/internal/dto"
)

type Repository interface {
	CreateWallet(walletName string) error
	IncreaseWalletBalance(walletName string, amount uint64) error

	RunWithTransaction(f func(tx *sqlx.Tx) error) error
	GetWalletsForUpdateTx(tx *sqlx.Tx, walletNames []string) ([]dto.Wallet, error)
	DecreaseWalletBalanceTx(tx *sqlx.Tx, walletName string, amount uint64) error
	IncreaseWalletBalanceTx(tx *sqlx.Tx, walletName string, amount uint64) error
}
