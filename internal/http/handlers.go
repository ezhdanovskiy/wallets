package http

import (
	"encoding/json"
	"net/http"

	"github.com/ezhdanovskiy/wallets/internal/dto"
)

func (s *Server) createWallet(w http.ResponseWriter, r *http.Request) {
	var wallet dto.Wallet
	if err := json.NewDecoder(r.Body).Decode(&wallet); err != nil {
		s.writeResponse(w, http.StatusBadRequest, err)
		return
	}

	err := s.svc.CreateWallet(wallet)
	if err != nil {
		s.writeResponse(w, http.StatusInternalServerError, err)
		return
	}

	s.writeResponse(w, http.StatusOK, nil)
}
