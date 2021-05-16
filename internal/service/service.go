// Package service contains the business logic for wallets application.
package service

import (
	"fmt"
	"net/http"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"

	"github.com/ezhdanovskiy/wallets/internal/dto"
	"github.com/ezhdanovskiy/wallets/internal/httperr"
)

type Service struct {
	log  *zap.SugaredLogger
	repo Repository
}

var ErrDatabase = httperr.New(http.StatusInternalServerError, "database error")

func NewService(logger *zap.SugaredLogger, repo Repository) *Service {
	return &Service{
		log:  logger,
		repo: repo,
	}
}

// CreateWallet creates new wallet.
func (s *Service) CreateWallet(wallet dto.CreateWalletRequest) error {
	err := s.repo.CreateWallet(wallet.Name)
	if err != nil {
		return ErrDatabase.Wrap(err)
	}
	return nil
}

// IncreaseWalletBalance increases wallet balance.
func (s *Service) IncreaseWalletBalance(deposit dto.Deposit) error {
	wallet, err := s.repo.GetWallet(deposit.Wallet)
	if err != nil {
		return ErrDatabase.Wrap(err)
	}
	if wallet == nil {
		return httperr.Wrap(err, http.StatusNotFound, "wallet not found")
	}

	err = s.repo.IncreaseWalletBalance(deposit.Wallet, deposit.Amount.GetInt())
	if err != nil {
		return ErrDatabase.Wrap(err)
	}
	return nil
}

func (s *Service) Transfer(transfer dto.Transfer) error {
	return s.repo.RunWithTransaction(func(tx *sqlx.Tx) error {
		wallets, err := s.repo.GetWalletsForUpdateTx(tx, []string{transfer.WalletFrom, transfer.WalletTo})
		if err != nil {
			return ErrDatabase.Wrap(fmt.Errorf("get wallets for update: %w", err))
		}

		if len(wallets) < 2 {
			if len(wallets) == 0 {
				return httperr.New(http.StatusNotFound, "wallets not found")
			}
			if wallets[0].Name == transfer.WalletFrom {
				return httperr.New(http.StatusNotFound, "wallet_to not found")
			}
			return httperr.New(http.StatusNotFound, "wallet_from not found")
		}

		for _, w := range wallets {
			if w.Name == transfer.WalletFrom {
				if w.Balance < transfer.Amount.GetInt() {
					return httperr.New(http.StatusUnprocessableEntity, "not enough money")
				}
				break
			}
		}

		err = s.repo.TransferTx(tx, transfer.WalletFrom, transfer.WalletTo, transfer.Amount.GetInt())
		if err != nil {
			return ErrDatabase.Wrap(fmt.Errorf("transfer: %w", err))
		}

		return nil
	})
}
