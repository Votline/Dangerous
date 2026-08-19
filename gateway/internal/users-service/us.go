// Package usersservice us.go provides implementation
// 'Service' interface for users-service
package usersservice

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"gateway/internal/security"

	pb "github.com/Votline/Dangerous/protos/generated-users"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type UsersService struct {
	name       string
	jwtSecret  string
	log        *zap.Logger
	ctxTimeout time.Duration
	jman       security.JWTSecurity
	conn       *grpc.ClientConn
	client     pb.UsersServiceClient
}

func (s *UsersService) Init(ctxTimeout time.Duration, mux *http.ServeMux, log *zap.Logger) error {
	const op = "users_service.Init"

	log.Debug("Creating users-service",
		zap.String("op", op))

	var jman security.JWTManager
	if err := jman.Init(); err != nil {
		return fmt.Errorf("%s: jwtmanager init: %w", op, err)
	}

	conn, err := grpc.NewClient(
		os.Getenv("USERS_SERVICE_HOST")+":"+os.Getenv("USERS_SERVICE_PORT"),
		grpc.WithInsecure())
	if err != nil {
		return fmt.Errorf("%s: new client: %w", op, err)
	}

	s.name = "users_service"
	s.log = log
	s.conn = conn
	s.jman = &jman
	s.ctxTimeout = ctxTimeout
	s.client = pb.NewUsersServiceClient(conn)

	s.registerRoutes(mux)

	return nil
}

func (s *UsersService) Shutdown(ctx context.Context) error {
	const op = "users_service.Shutdown"

	if err := s.conn.Close(); err != nil {
		return fmt.Errorf("%s: grpc conn close: %w", op, err)
	}

	return nil
}

func (s *UsersService) GetName() string {
	return s.name
}

func (s *UsersService) registerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /register", s.Register)
}
