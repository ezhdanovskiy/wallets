// Package service contains the business logic for wallets application.
package service

import (
	"fmt"
	"net/http"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"

	"github.com/ezhdanovskiy/wallets/internal/consts"
	"github.com/ezhdanovskiy/wallets/internal/dto"
	"github.com/ezhdanovskiy/wallets/internal/httperr"
)

// Service implements the business logic for wallets application.
type Service struct {
	log  *zap.SugaredLogger
	repo Repository
}

// NewService creates a service instance.
func NewService(logger *zap.SugaredLogger, repo Repository) *Service {
	return &Service{
		log:  logger,
		repo: repo,
	}
}

// CreateWallet creates new wallet.
func (s *Service) CreateWallet(wallet dto.CreateWalletRequest) error {
	if wallet.Name == "" {
		return ErrEmptyWalletName
	}

	err := s.repo.CreateWallet(wallet.Name)
	if err != nil {
		return ErrDatabase.Wrap(err)
	}
	return nil
}

// IncreaseWalletBalance increases wallet balance.
func (s *Service) IncreaseWalletBalance(deposit dto.Deposit) error {
	if deposit.Wallet == "" {
		return ErrEmptyWalletName
	}
	if deposit.Amount <= 0 {
		return ErrNotPositiveAmount
	}

	wallet, err := s.repo.GetWallet(deposit.Wallet)
	if err != nil {
		return ErrDatabase.Wrap(err)
	}
	if wallet == nil {
		return ErrWalletNotFound
	}

	err = s.repo.IncreaseWalletBalance(deposit.Wallet, deposit.Amount.GetInt())
	if err != nil {
		return ErrDatabase.Wrap(err)
	}
	return nil
}

// Transfer transfers money from one wallet to another.
func (s *Service) Transfer(transfer dto.Transfer) error {
	if transfer.WalletFrom == "" {
		return ErrEmptyWalletFrom
	}
	if transfer.WalletTo == "" {
		return ErrEmptyWalletTo
	}
	if transfer.WalletFrom == transfer.WalletTo {
		return ErrSameWallets
	}
	if transfer.Amount <= 0 {
		return ErrNotPositiveAmount
	}

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
				return httperr.New(http.StatusNotFound, "%s not found", transfer.WalletTo)
			}
			return httperr.New(http.StatusNotFound, "%s not found", transfer.WalletFrom)
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

// GetOperations provides operations for the specified wallet according to filtering parameters.
func (s *Service) GetOperations(filter dto.OperationsFilter) ([]dto.Operation, error) {
	if filter.Wallet == "" {
		return nil, ErrEmptyWalletName
	}
	if filter.Type != "" && filter.Type != consts.OperationTypeDeposit && filter.Type != consts.OperationTypeWithdrawal {
		return nil, ErrUnsupportedOperationType
	}
	if filter.StartDate < 0 {
		return nil, ErrNegativeStartDate
	}
	if filter.EndDate < 0 {
		return nil, ErrNegativeEndDate
	}
	if filter.Limit < 0 {
		return nil, ErrNotPositiveLimit
	}
	if filter.Limit == 0 {
		filter.Limit = consts.OperationsLimitDefault
	}
	if filter.Offset < 0 {
		return nil, ErrNegativeOffset
	}

	operations, err := s.repo.GetOperations(filter)
	if err != nil {
		return nil, ErrDatabase.Wrap(err)
	}
	return operations, nil
}
