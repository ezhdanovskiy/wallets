// Package application runs the required components depending on the parameters.
package application

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/ezhdanovskiy/wallets/internal/config"
	"github.com/ezhdanovskiy/wallets/internal/httpsrv"
	"github.com/ezhdanovskiy/wallets/internal/repository"
	"github.com/ezhdanovskiy/wallets/internal/service"
)

// Application contains all element of application.
type Application struct {
	log *zap.SugaredLogger
	cfg *config.Config
	svc *service.Service

	httpServer *httpsrv.Server

	ctx    context.Context
	cancel context.CancelFunc
}

// NewApplication creates instance of Application with configured components.
func NewApplication() (*Application, error) {
	cfg, err := config.NewConfig()
	if err != nil {
		return nil, fmt.Errorf("new config: %w", err)
	}

	log, err := newLogger(cfg.LogLevel, cfg.LogEncoding)
	if err != nil {
		return nil, fmt.Errorf("new logger: %w", err)
	}
	log.Debugf("cfg: %+v", cfg)

	repo, err := repository.NewRepo(log, cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)
	if err != nil {
		return nil, fmt.Errorf("new repo: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Application{
		log:    log,
		cfg:    cfg,
		svc:    service.NewService(log, repo),
		ctx:    ctx,
		cancel: cancel,
	}, nil
}

// Run runs configured components.
func (a *Application) Run() error {
	a.log.Info("Run application")

	a.httpServer = httpsrv.NewServer(a.log, a.cfg.HttpPort, a.svc)

	a.log.Infof("Run HTTP server on port %v", a.cfg.HttpPort)
	err := a.httpServer.Run()
	if err != nil {
		return fmt.Errorf("HTTP server run: %w", err)
	}
	a.log.Info("HTTP server stopped")

	a.log.Info("Application stopped")
	return nil
}

// Stop terminates configured components.
func (a *Application) Stop() {
	if a.cancel != nil {
		a.cancel()
	}
	if a.httpServer != nil {
		a.log.Info("Stopping HTTP server")
		a.httpServer.Shutdown()
	}
}
