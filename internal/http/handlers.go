package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/ezhdanovskiy/wallets/csv"
	"github.com/ezhdanovskiy/wallets/internal/dto"
	"github.com/ezhdanovskiy/wallets/internal/httperr"
)

var ErrBodyDecode = httperr.New(http.StatusBadRequest, "failed to decode body")

const (
	LimitDefault = 20
	LimitMax     = 1000
)

func (s *Server) createWallet(w http.ResponseWriter, r *http.Request) {
	s.log.Debug("createWallet")

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

func (s *Server) getOperations(w http.ResponseWriter, r *http.Request) {
	filter := dto.OperationsFilter{
		Wallet: r.URL.Query().Get("wallet"),
		Type:   r.URL.Query().Get("type"),
	}

	if filter.Wallet == "" {
		s.writeErrorResponse(w, httperr.New(http.StatusBadRequest, "empty wallet parameter"))
		return
	}

	if startDate := r.URL.Query().Get("start_date"); startDate != "" {
		i, err := strconv.ParseInt(startDate, 10, 64)
		if err != nil {
			s.writeErrorResponse(w, httperr.Wrap(err, http.StatusBadRequest, "failed to parse start_date"))
			return
		}
		filter.StartDate = i
	}

	if endDate := r.URL.Query().Get("end_date"); endDate != "" {
		i, err := strconv.ParseInt(endDate, 10, 64)
		if err != nil {
			s.writeErrorResponse(w, httperr.Wrap(err, http.StatusBadRequest, "failed to parse end_date"))
			return
		}
		filter.EndDate = i
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		limit, err := strconv.ParseInt(limitStr, 10, 64)
		if err != nil {
			s.writeErrorResponse(w, httperr.Wrap(err, http.StatusBadRequest, "failed to parse limit"))
			return
		}
		if limit < 1 || limit > LimitMax {
			s.writeErrorResponse(w, httperr.Wrap(err, http.StatusBadRequest, "wrong limit, it have to be in [1, 100]"))
			return
		}
		filter.Limit = limit
	} else {
		filter.Limit = LimitDefault
	}

	if offset := r.URL.Query().Get("offset"); offset != "" {
		i, err := strconv.ParseInt(offset, 10, 64)
		if err != nil {
			s.writeErrorResponse(w, httperr.Wrap(err, http.StatusBadRequest, "failed to parse offset"))
			return
		}
		filter.Offset = i
	}

	operations, err := s.svc.GetOperations(filter)
	if err != nil {
		s.writeErrorResponse(w, err)
		return
	}

	if r.URL.Query().Get("format") == "csv" {
		data, err := csv.ConvertOperations(operations)
		if err != nil {
			s.writeErrorResponse(w, err)
			return
		}
		s.writeResponse(w, http.StatusOK, data)
		return
	}

	s.writeResponse(w, http.StatusOK, operations)
}
