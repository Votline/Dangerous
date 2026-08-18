// Package main provides users-service grpc methods
package main

import (
	"net"

	"usrsrv/internal/utils"

	pb "github.com/Votline/Dangerous/protos/generated-users"
	"google.golang.org/grpc"

	"go.uber.org/zap"
)

type usersserver struct {
	log *zap.Logger
	pb.UnimplementedUsersServiceServer
}

func main() {
	log, _ := zap.NewDevelopment()

	addr := utils.GetEnvString("USERS_SERVICE_ADDR", ":50051")
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal("failed to create listen", zap.Error(err))
	}

	srv := usersserver{log: log}
	gsrv := grpc.NewServer()
	pb.RegisterUsersServiceServer(gsrv, &srv)

	log.Debug("Starting users-service...", zap.String("addr", addr))
	if err := gsrv.Serve(lis); err != nil {
		log.Fatal("failed to start grpc", zap.Error(err))
	}
}
