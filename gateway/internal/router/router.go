// Package router router.go contains implementation 'Server'
// interface by 'HTTPServer' struct
// Creates and manages the main http server
package router

import (
	"context"
	"fmt"
	"net/http"

	"go.uber.org/zap"
)

type Server interface {
	Init() error
	Start() error
	Shutdown(ctx context.Context) error
}

type HTTPServer struct {
	srv *http.Server
	log *zap.Logger
}

func (s *HTTPServer) Init(log *zap.Logger) error {
	mux := http.NewServeMux()
	s.srv = &http.Server{
		Handler: mux,
	}
	s.log = log
	return nil
}

func (s *HTTPServer) Start() error {
	const op = "router.Start"
	if err := s.srv.ListenAndServe(); err != nil {
		return fmt.Errorf("%s: listen and serve: %w", op, err)
	}
	return nil
}

func (s *HTTPServer) Shutdown(ctx context.Context) error {
	const op = "router.Shutdown"

	if err := s.srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("%s: shutdown: %w", op, err)
	}
	return nil
}
