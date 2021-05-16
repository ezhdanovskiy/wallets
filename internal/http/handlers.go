package http

import (
	"encoding/json"
	"net/http"

	"github.com/ezhdanovskiy/wallets/internal/dto"
	"github.com/ezhdanovskiy/wallets/internal/httperr"
)

var ErrBodyDecode = httperr.New(http.StatusBadRequest, "failed to decode body")

func (s *Server) createWallet(w http.ResponseWriter, r *http.Request) {
	var wallet dto.CreateWalletRequest
	if err := json.NewDecoder(r.Body).Decode(&wallet); err != nil {
		s.writeErrorResponse(w, ErrBodyDecode.Wrap(err))
		return
	}

	err := s.svc.CreateWallet(wallet)
	if err != nil {
		s.writeErrorResponse(w, err)
		return
	}

	s.writeResponse(w, http.StatusOK, nil)
}

func (s *Server) deposit(w http.ResponseWriter, r *http.Request) {
	var deposit dto.Deposit
	if err := json.NewDecoder(r.Body).Decode(&deposit); err != nil {
		s.writeErrorResponse(w, ErrBodyDecode.Wrap(err))
		return
	}

	err := s.svc.IncreaseWalletBalance(deposit)
	if err != nil {
		s.writeErrorResponse(w, err)
		return
	}

	s.writeResponse(w, http.StatusOK, nil)
}

func (s *Server) transfer(w http.ResponseWriter, r *http.Request) {
	var transfer dto.Transfer
	if err := json.NewDecoder(r.Body).Decode(&transfer); err != nil {
		s.writeErrorResponse(w, ErrBodyDecode.Wrap(err))
		return
	}

	err := s.svc.Transfer(transfer)
	if err != nil {
		s.writeErrorResponse(w, err)
		return
	}

	s.writeResponse(w, http.StatusOK, nil)
}
