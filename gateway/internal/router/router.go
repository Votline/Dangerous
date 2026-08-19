// Package router router.go contains implementation 'Server'
// interface by 'HTTPServer' struct
// Creates and manages the main http server
package router

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"gateway/internal/services"
	usersservice "gateway/internal/users-service"
	"gateway/internal/utils"

	"go.uber.org/zap"
)

type Server interface {
	Init() error
	Start() error
	Shutdown(ctx context.Context) error
}

type HTTPServer struct {
	srv  *http.Server
	log  *zap.Logger
	svcs []services.Service
}

func (s *HTTPServer) Init(log *zap.Logger) error {
	addr := ":" + utils.GetEnvString("GATEWAY_PORT", "8080")

	mux := http.NewServeMux()
	s.srv = &http.Server{
		Addr:    addr,
		Handler: mux,
	}
	s.log = log

	s.registerServices(mux)

	return nil
}

func (s *HTTPServer) Start() error {
	const op = "router.Start"

	s.log.Debug("Starting...",
		zap.String("op", op),
		zap.String("addr", s.srv.Addr))
	if err := s.srv.ListenAndServe(); err != nil {
		return fmt.Errorf("%s: listen and serve: %w", op, err)
	}
	return nil
}

func (s *HTTPServer) Shutdown(ctx context.Context) error {
	const op = "router.Shutdown"

	s.log.Debug("Shutdowning http server...", zap.String("op", op))
	if err := s.srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("%s: shutdown: %w", op, err)
	}

	s.log.Debug("Shutdowning services...", zap.String("op", op))

	for _, svc := range s.svcs {
		if err := svc.Shutdown(ctx); err != nil {
			s.log.Error("Shutdown failed",
				zap.String("op", op),
				zap.String("name", svc.GetName()),
				zap.Error(err))
		}
		s.log.Error("Shutdowned",
			zap.String("op", op),
			zap.String("name", svc.GetName()))
	}

	return nil
}

func (s *HTTPServer) registerServices(mux *http.ServeMux) error {
	const op = "router.registerServices"

	ctxTimeout := time.Duration(utils.GetEnvInt("CTX_TIMEOUT", 10)) * time.Second

	var us usersservice.UsersService
	if err := us.Init(ctxTimeout, mux, s.log); err != nil {
		return fmt.Errorf("%s: users-service init: %w", op, err)
	}

	s.svcs = make([]services.Service, 0, 1)
	s.svcs = append(s.svcs, &us)

	return nil
}
