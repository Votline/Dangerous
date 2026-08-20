// Package main provides marks-service grpc methods
package main

import (
	"context"
	"net"

	"mrksrv/internal/utils"

	pb "github.com/Votline/Dangerous/protos/generated-marks"
	"google.golang.org/grpc"

	"go.uber.org/zap"
)

type marksserver struct {
	log *zap.Logger
	pb.UnimplementedMarksServiceServer
}

func main() {
	log, _ := zap.NewDevelopment()

	addr := ":" + utils.GetEnvString("MARKS_SERVICE_PORT", "50052")
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal("failed to create listen", zap.Error(err))
	}

	srv := marksserver{log: log}
	gsrv := grpc.NewServer()
	pb.RegisterMarksServiceServer(gsrv, &srv)

	log.Debug("Starting marks-service...", zap.String("addr", addr))
	if err := gsrv.Serve(lis); err != nil {
		log.Fatal("failed to start grpc", zap.Error(err))
	}
}

func (s *marksserver) New(ctx context.Context, req *pb.NewReq) (*pb.NewRes, error) {
	const op = "marksserver.New"

	return nil, nil
}

func (s *marksserver) Get(ctx context.Context, req *pb.GetReq) (*pb.GetRes, error) {
	const op = "marksserver.Get"

	return nil, nil
}
