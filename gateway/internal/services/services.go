// Package services provides the interface for the services
// and help-funcs for grpc call
package services

import (
	"context"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type Service interface {
	GetName() string
	Init(ctxTimeout time.Duration, mux *http.ServeMux, log *zap.Logger) error
	Shutdown(ctx context.Context) error
}
