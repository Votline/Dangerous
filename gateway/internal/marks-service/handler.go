// Package marksservice handler.go provides handlers for marks service.
package marksservice

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"gateway/internal/services"

	pb "github.com/Votline/Dangerous/protos/generated-marks"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Request struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Comment   string  `json:"comment"`
}

type AdditionalInfo struct {
	Nickname string `json:"nickname"`
	Comment  string `json:"comment"`
}

func (s *MarksService) New(w http.ResponseWriter, r *http.Request) {
	const op = "marks_service.New"

	reqTrace := uuid.NewString()

	s.log.Debug("New marks request",
		zap.String("op", op),
		zap.String("reqTrace", reqTrace))

	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("%s: decode request: %s", op, err.Error()), http.StatusBadRequest)
		return
	}

	uinfo, err := s.jman.GetUserInfo(r)
	if err != nil {
		http.Error(w, fmt.Sprintf("%s: get user info: %s", op, err.Error()), http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), s.ctxTimeout)
	defer cancel()

	if _, err := services.CallRPC(s.cb, func() (*pb.NewRes, error) {
		return s.client.New(ctx, &pb.NewReq{
			Nickname:     uinfo.Nickname,
			Comment:      req.Comment,
			Latitude:     req.Latitude,
			Longitude:    req.Longitude,
			RequestTrace: reqTrace,
		})
	}); err != nil {
		http.Error(w, fmt.Sprintf("%s: rpc call: %s", op, err.Error()), http.StatusInternalServerError)
		return
	}

	s.log.Debug("Successfully registed",
		zap.String("op", op),
		zap.String("reqTrace", reqTrace))
}

func (s *MarksService) Get(w http.ResponseWriter, r *http.Request) {
	const op = "marks_service.Get"

	reqTrace := uuid.NewString()

	s.log.Debug("Get request",
		zap.String("op", op),
		zap.String("reqTrace", reqTrace))

	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("%s: decode request: %s", op, err.Error()), http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), s.ctxTimeout)
	defer cancel()

	res, err := services.CallRPC(s.cb, func() (*pb.GetRes, error) {
		return s.client.Get(ctx, &pb.GetReq{
			Latitude:     req.Latitude,
			Longitude:    req.Longitude,
			RequestTrace: reqTrace,
		})
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("%s: rpc call: %s", op, err.Error()), http.StatusInternalServerError)
		return
	}

	s.log.Debug("Successfully getted marks",
		zap.String("op", op),
		zap.String("reqTrace", reqTrace))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res.Comments)
}
