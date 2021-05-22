package http

import (
	"github.com/ezhdanovskiy/wallets/internal/dto"
)

// Service describes the service methods required for the server.
type Service interface {
	CreateWallet(dto.CreateWalletRequest) error
	IncreaseWalletBalance(dto.Deposit) error
	Transfer(dto.Transfer) error
	GetOperations(dto.OperationsFilter) ([]dto.Operation, error)
}
