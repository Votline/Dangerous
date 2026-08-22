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

	s.log.Debug("Register request",
		zap.String("op", op),
		zap.String("reqTrace", reqTrace))

	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("%s: decode request: %s", op, err.Error()), http.StatusBadRequest)
		return
	}

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

	cookie := &http.Cookie{
		Name:     "token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   3600,
	}
	http.SetCookie(w, cookie)
}

func (s *UsersService) Login(w http.ResponseWriter, r *http.Request) {
	const op = "users_service.Login"

	reqTrace := uuid.NewString()

	s.log.Debug("Login request",
		zap.String("op", op),
		zap.String("reqTrace", reqTrace))

	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("%s: decode request: %s", op, err.Error()), http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), s.ctxTimeout)
	defer cancel()

	if _, err := s.client.Login(ctx, &pb.LogReq{
		Nickname:     req.Nickname,
		Password:     req.Password,
		RequestTrace: reqTrace,
	}); err != nil {
		http.Error(w, fmt.Sprintf("%s: rpc call: %s", op, err.Error()), http.StatusInternalServerError)
		return
	}

	s.log.Debug("Successfully logged in",
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

	cookie := &http.Cookie{
		Name:     "token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   3600,
	}
	http.SetCookie(w, cookie)
}

func (s *UsersService) Delete(w http.ResponseWriter, r *http.Request) {
	const op = "users_service.Delete"

	reqTrace := uuid.NewString()

	s.log.Debug("Delete request",
		zap.String("op", op),
		zap.String("reqTrace", reqTrace))

	password := r.Header.Get("X-Confirm-Password")
	if password == "" {
		http.Error(w, fmt.Sprintf("%s: get password: password confirmation required", op), http.StatusBadRequest)
		return
	}

	cookie, err := r.Cookie("token")
	if err != nil {
		http.Error(w, fmt.Sprintf("%s: get cookie: %s", op, err.Error()), http.StatusBadRequest)
		return
	}

	uinfo, err := s.jman.ExtractClaims(cookie.Value)
	if err != nil {
		http.Error(w, fmt.Sprintf("%s: extract claims: %s", op, err.Error()), http.StatusBadRequest)
		return
	}
	nickname := uinfo.Nickname

	s.log.Debug("Getted user info",
		zap.String("op", op),
		zap.String("reqTrace", reqTrace))

	ctx, cancel := context.WithTimeout(r.Context(), s.ctxTimeout)
	defer cancel()

	if _, err := s.client.Delete(ctx, &pb.DelReq{
		Nickname:     nickname,
		Password:     password,
		RequestTrace: reqTrace,
	}); err != nil {
		http.Error(w, fmt.Sprintf("%s: rpc call: %s", op, err.Error()), http.StatusInternalServerError)
		return
	}

	s.log.Debug("Successfully deleted",
		zap.String("op", op),
		zap.String("reqTrace", reqTrace))

	w.WriteHeader(http.StatusNoContent)
}
