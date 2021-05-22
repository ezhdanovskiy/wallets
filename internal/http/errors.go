package http

import (
	"net/http"

	"github.com/ezhdanovskiy/wallets/internal/httperr"
)

var (
	ErrBodyDecode = httperr.New(http.StatusBadRequest, "failed to decode body")
)
