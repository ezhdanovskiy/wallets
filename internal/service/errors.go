package service

import (
	"net/http"

	"github.com/ezhdanovskiy/wallets/internal/httperr"
)

var (
	ErrDatabase                 = httperr.New(http.StatusInternalServerError, "database error")
	ErrEmptyWalletFrom          = httperr.New(http.StatusBadRequest, "empty wallet_from")
	ErrEmptyWalletName          = httperr.New(http.StatusBadRequest, "empty wallet name")
	ErrEmptyWalletTo            = httperr.New(http.StatusBadRequest, "empty wallet_to")
	ErrSameWallets              = httperr.New(http.StatusBadRequest, "same wallets")
	ErrNegativeEndDate          = httperr.New(http.StatusBadRequest, "end_date can't be negative")
	ErrNegativeOffset           = httperr.New(http.StatusBadRequest, "offset can't be negative")
	ErrNegativeStartDate        = httperr.New(http.StatusBadRequest, "start_date can't be negative")
	ErrNotPositiveAmount        = httperr.New(http.StatusBadRequest, "amount must be positive")
	ErrNotPositiveLimit         = httperr.New(http.StatusBadRequest, "limit must be positive")
	ErrUnsupportedOperationType = httperr.New(http.StatusBadRequest, "unsupported operation type")
	ErrWalletNotFound           = httperr.New(http.StatusBadRequest, "wallet not found")
)
