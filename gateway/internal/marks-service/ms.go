// Package marksservice ms.go provides implementation
// 'Service' interface for marks-service
package marksservice

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"gateway/internal/security"

	pb "github.com/Votline/Dangerous/protos/generated-marks"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type MarksService struct {
	name       string
	jwtSecret  string
	log        *zap.Logger
	ctxTimeout time.Duration
	jman       security.JWTSecurity
	conn       *grpc.ClientConn
	client     pb.MarksServiceClient
}

func (s *MarksService) Init(ctxTimeout time.Duration, mux *http.ServeMux, log *zap.Logger) error {
	const op = "marks_service.Init"

	log.Debug("Creating marks-service",
		zap.String("op", op))

	var jman security.JWTManager
	if err := jman.Init(); err != nil {
		return fmt.Errorf("%s: jwtmanager init: %w", op, err)
	}

	conn, err := grpc.NewClient(
		os.Getenv("MARKS_SERVICE_HOST")+":"+os.Getenv("MARKS_SERVICE_PORT"),
		grpc.WithInsecure())
	if err != nil {
		return fmt.Errorf("%s: new client: %w", op, err)
	}

	s.name = "marks_service"
	s.log = log
	s.conn = conn
	s.jman = &jman
	s.ctxTimeout = ctxTimeout
	s.client = pb.NewMarksServiceClient(conn)

	s.registerRoutes(mux)

	return nil
}

func (s *MarksService) Shutdown(ctx context.Context) error {
	const op = "marks_service.Shutdown"

	if err := s.conn.Close(); err != nil {
		return fmt.Errorf("%s: grpc conn close: %w", op, err)
	}

	return nil
}

func (s *MarksService) GetName() string {
	return s.name
}

func (s *MarksService) registerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/marks/new", s.New)
	mux.HandleFunc("POST /api/marks/get", s.Get)
}
