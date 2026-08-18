// Package main provides users-service grpc methods
package main

import (
	"context"
	"fmt"
	"net"

	"usrsrv/internal/db"
	"usrsrv/internal/security"
	"usrsrv/internal/utils"

	pb "github.com/Votline/Dangerous/protos/generated-users"
	"google.golang.org/grpc"

	"go.uber.org/zap"
)

type usersserver struct {
	udb db.DB
	log *zap.Logger
	pb.UnimplementedUsersServiceServer
}

func main() {
	log, _ := zap.NewDevelopment()

	addr := ":" + utils.GetEnvString("USERS_SERVICE_PORT", "50051")
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal("failed to create listen", zap.Error(err))
	}

	udb, err := db.NewUDB(log)
	if err != nil {
		log.Fatal("failed to create db", zap.Error(err))
	}

	srv := usersserver{log: log, udb: udb}
	gsrv := grpc.NewServer()
	pb.RegisterUsersServiceServer(gsrv, &srv)

	log.Debug("Starting users-service...", zap.String("addr", addr))
	if err := gsrv.Serve(lis); err != nil {
		log.Fatal("failed to start grpc", zap.Error(err))
	}
}

func (s *usersserver) Register(ctx context.Context, req *pb.RegReq) (*pb.RegRes, error) {
	const op = "usersserver.Register"

	nickname := req.GetNickname()
	rawPassword := req.GetPassword()
	reqTrace := req.GetRequestTrace()

	s.log.Info("Register request",
		zap.String("op", op),
		zap.String("reqTrace", reqTrace),
		zap.String("nickname", nickname))

	hashPassword, err := security.Hash(rawPassword)
	if err != nil {
		return nil, fmt.Errorf("%s: hash password: %w", op, err)
	}

	s.log.Info("Successfully hashed password",
		zap.String("op", op),
		zap.String("reqTrace", reqTrace))

	if err := s.udb.Register(nickname, hashPassword, reqTrace, ctx); err != nil {
		return nil, fmt.Errorf("%s: register to db: %w", op, err)
	}

	s.log.Info("Successfully registred user",
		zap.String("op", op),
		zap.String("reqTrace", reqTrace))

	return &pb.RegRes{}, nil
}

func (s *usersserver) Login(ctx context.Context, req *pb.LogReq) (*pb.LogRes, error) {
	const op = "usersserver.Login"

	nickname := req.GetNickname()
	rawPassword := req.GetPassword()
	reqTrace := req.GetRequestTrace()

	s.log.Info("Login request",
		zap.String("op", op),
		zap.String("reqTrace", reqTrace),
		zap.String("nickname", nickname))

	gettedPswd, err := s.udb.Get(nickname, reqTrace, ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: login to db: %w", op, err)
	}

	s.log.Info("Getted user",
		zap.String("op", op),
		zap.String("reqTrace", reqTrace))

	if err := security.Check(rawPassword, gettedPswd); err != nil {
		return nil, fmt.Errorf("%s: check password: %w", op, err)
	}

	s.log.Info("Security confirmed",
		zap.String("op", op),
		zap.String("reqTrace", reqTrace))

	return &pb.LogRes{}, nil
}

func (s *usersserver) Delete(ctx context.Context, req *pb.DelReq) (*pb.DelRes, error) {
	const op = "usersserver.Delete"

	nickname := req.GetNickname()
	rawPassword := req.GetPassword()
	reqTrace := req.GetRequestTrace()

	s.log.Info("Delete request",
		zap.String("op", op),
		zap.String("reqTrace", reqTrace),
		zap.String("nickname", nickname))

	if _, err := s.Login(ctx, &pb.LogReq{
		Nickname:     nickname,
		Password:     rawPassword,
		RequestTrace: reqTrace,
	}); err != nil {
		return nil, fmt.Errorf("%s: authentication failed: %w", op, err)
	}

	s.log.Info("Authentication confirmed",
		zap.String("op", op),
		zap.String("reqTrace", reqTrace))

	if err := s.udb.Delete(nickname, reqTrace, ctx); err != nil {
		return nil, fmt.Errorf("%s: delete from db: %w", op, err)
	}

	s.log.Info("Successfully deleted",
		zap.String("op", op),
		zap.String("reqTrace", reqTrace))

	return nil, nil
}
