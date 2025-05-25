package http

import (
	"github.com/ezhdanovskiy/wallets/internal/dto"
)

//go:generate mockgen -source=dependencies.go -destination=mocks/service_mock.go -package=mocks

// Service describes the service methods required for the server.
type Service interface {
	CreateWallet(dto.CreateWalletRequest) error
	IncreaseWalletBalance(dto.Deposit) error
	Transfer(dto.Transfer) error
	GetOperations(dto.OperationsFilter) ([]dto.Operation, error)
}
