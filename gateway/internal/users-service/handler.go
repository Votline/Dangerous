// Package usersservice handler.go provides handlers for users service.
package usersservice

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	pb "github.com/Votline/Dangerous/protos/generated-users"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Request struct {
	Nickname string `json:"nickname"`
	Password string `json:"password"`
}

func (s *UsersService) Register(w http.ResponseWriter, r *http.Request) {
	const op = "users_service.Register"

	reqTrace := uuid.NewString()

	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("%s: decode request: %s", op, err.Error()), http.StatusBadRequest)
		return
	}

	s.log.Debug("Register request",
		zap.String("op", op),
		zap.String("reqTrace", reqTrace))

	ctx, cancel := context.WithTimeout(r.Context(), s.ctxTimeout)
	defer cancel()

	if _, err := s.client.Register(ctx, &pb.RegReq{
		Nickname:     req.Nickname,
		Password:     req.Password,
		RequestTrace: reqTrace,
	}); err != nil {
		http.Error(w, fmt.Sprintf("%s: rpc call: %s", op, err.Error()), http.StatusInternalServerError)
		return
	}

	s.log.Debug("Successfully registed",
		zap.String("op", op),
		zap.String("reqTrace", reqTrace))

	token, err := s.jman.GenerateToken(req.Nickname)
	if err != nil {
		http.Error(w, fmt.Sprintf("%s: generate token: %s", op, err.Error()), http.StatusInternalServerError)
		return
	}

	s.log.Debug("Successfully generated jwt",
		zap.String("op", op),
		zap.String("reqTrace", reqTrace))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"token": token,
	})
}
