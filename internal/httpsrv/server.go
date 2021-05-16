// Package httpsrv contains the HTTP server and associated endpoint handlers.
package httpsrv

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	"github.com/ezhdanovskiy/wallets/internal/dto"
	"github.com/ezhdanovskiy/wallets/internal/httperr"
)

type Server struct {
	log        *zap.SugaredLogger
	httpPort   int
	httpServer *http.Server
	svc        Service
}

type Service interface {
	CreateWallet(wallet dto.CreateWalletRequest) error
	IncreaseWalletBalance(deposit dto.Deposit) error
	Transfer(transfer dto.Transfer) error
}

func NewServer(logger *zap.SugaredLogger, httpPort int, svc Service) *Server {
	return &Server{
		log:      logger,
		httpPort: httpPort,
		svc:      svc,
	}
}

func (s *Server) Run() error {
	router := chi.NewMux()

	router.Handle(
		"/metrics",
		promhttp.Handler(),
	)

	router.Route("/v1", s.GetV1ApiRouters())

	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.httpPort),
		Handler: router,
	}

	err := s.httpServer.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("start http server: %w", err)
	}

	return nil
}

func (s *Server) GetV1ApiRouters() func(chi.Router) {
	return func(r chi.Router) {
		r.Post("/wallets", s.createWallet)
		r.Post("/wallets/deposit", s.deposit)
		r.Post("/wallets/transfer", s.transfer)
	}
}

func (s *Server) Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.log.Errorf("http server shutdown: %s", err)
	}
}

type Resp struct {
	Error string      `json:"error,omitempty"`
	Data  interface{} `json:"data,omitempty"`
}

func (s *Server) writeResponse(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("content-type", "application/json; charset=utf-8")
	w.WriteHeader(code)

	if payload == nil {
		return
	}

	data, err := json.Marshal(Resp{Data: payload})
	if err != nil {
		s.log.With("error", err).Error("failed to marshal json")
		return
	}
	s.log.Debugf("Send response: %s", data)

	if _, err := w.Write(data); err != nil {
		s.log.Errorf("http response: %s", err.Error())
	}
}

func (s *Server) writeErrorResponse(w http.ResponseWriter, err error) {
	w.Header().Set("content-type", "application/json; charset=utf-8")
	if err == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	s.log.Error(err.Error())

	var resp Resp
	if e, ok := err.(*httperr.Error); ok {
		w.WriteHeader(e.StatusCode)
		resp.Error = e.Message
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		resp.Error = err.Error()
	}

	data, err := json.Marshal(resp)
	if err != nil {
		s.log.With("error", err).Error("failed to marshal json")
		return
	}

	if _, err := w.Write(data); err != nil {
		s.log.Errorf("http error response: %s", err.Error())
	}
	return
}
