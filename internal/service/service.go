// Package service contains the business logic for wallets application.
package service

import (
	"go.uber.org/zap"

	"github.com/ezhdanovskiy/wallets/internal/dto"
)

type Service struct {
	log  *zap.SugaredLogger
	repo Repository
}

type Repository interface {
	CreateWallet(wallet dto.Wallet) error
}

func NewService(logger *zap.SugaredLogger, repo Repository) *Service {
	return &Service{
		log:  logger,
		repo: repo,
	}
}

// CreateWallet creates new wallet.
func (s *Service) CreateWallet(wallet dto.Wallet) error {
	return s.repo.CreateWallet(wallet)
}
