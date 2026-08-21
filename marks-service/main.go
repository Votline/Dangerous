// Package main provides marks-service grpc methods
package main

import (
	"context"
	"fmt"
	"math"
	"net"

	"mrksrv/internal/db"
	"mrksrv/internal/utils"

	pb "github.com/Votline/Dangerous/protos/generated-marks"
	"google.golang.org/grpc"

	"go.uber.org/zap"
)

type marksserver struct {
	mdb db.DB
	log *zap.Logger
	pb.UnimplementedMarksServiceServer
	roundFactor float64
}

func main() {
	log, _ := zap.NewDevelopment()

	addr := ":" + utils.GetEnvString("MARKS_SERVICE_PORT", "50052")
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal("failed to create listen", zap.Error(err))
	}

	mdb, err := db.NewMDB(log)
	if err != nil {
		log.Fatal("failed to create mdb", zap.Error(err))
	}

	srv := marksserver{log: log, mdb: mdb}
	gsrv := grpc.NewServer()
	pb.RegisterMarksServiceServer(gsrv, &srv)

	srv.roundFactor = float64(utils.GetEnvInt("ROUND_FACTOR", 1000))

	log.Debug("Starting marks-service...", zap.String("addr", addr))
	if err := gsrv.Serve(lis); err != nil {
		log.Fatal("failed to start grpc", zap.Error(err))
	}
}

func (s *marksserver) New(ctx context.Context, req *pb.NewReq) (*pb.NewRes, error) {
	const op = "marksserver.New"

	reqTrace := req.GetRequestTrace()
	nickname := req.GetNickname()
	comment := req.GetComment()
	lat := req.GetLatitude()
	lng := req.GetLongitude()

	s.log.Debug("New request",
		zap.String("op", op),
		zap.String("reqTrace", reqTrace),
		zap.Float64("latitude", lat),
		zap.Float64("longitude", lng))

	if err := s.mdb.New(nickname, comment, reqTrace, lat, lng, ctx); err != nil {
		return nil, fmt.Errorf("%s: new mark: %w", op, err)
	}

	s.log.Debug("Successfully marked",
		zap.String("op", op),
		zap.String("reqTrace", reqTrace))

	return &pb.NewRes{}, nil
}

func (s *marksserver) Get(ctx context.Context, req *pb.GetReq) (*pb.GetRes, error) {
	const op = "marksserver.Get"

	reqTrace := req.GetRequestTrace()
	lat := math.Round(req.GetLatitude()*s.roundFactor) / s.roundFactor  // grid 3 decimal
	lng := math.Round(req.GetLongitude()*s.roundFactor) / s.roundFactor // grid 3 decimal

	s.log.Debug("New request",
		zap.String("op", op),
		zap.String("reqTrace", reqTrace),
		zap.Float64("latitude", lat),
		zap.Float64("longitude", lng))

	addInfo, err := s.mdb.Get(lat, lng, reqTrace, ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: get marks: %w", op, err)
	}

	s.log.Debug("Successfully marked",
		zap.String("op", op),
		zap.String("reqTrace", reqTrace))

	protoComments := make([]*pb.CommentItem, 0, len(addInfo))
	for _, c := range addInfo {
		protoComments = append(protoComments, &pb.CommentItem{
			Nickname: c.Nickname,
			Comment:  c.Comment,
		})
	}

	s.log.Debug("Successfully built comments",
		zap.String("op", op),
		zap.String("reqTrace", reqTrace))

	return &pb.GetRes{
		TotalReports: int32(len(addInfo)),
		Comments:     protoComments,
	}, nil
}
