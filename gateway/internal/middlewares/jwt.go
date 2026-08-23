package middlewares

import (
	"context"
	"fmt"
	"net/http"

	"gateway/internal/security"
)

type contextKey string

const UserCtxKey contextKey = "user_info"

type JWTMiddleware struct {
	jwtManager security.JWTSecurity
}

func NewJWTMiddleware(jwtManager security.JWTSecurity) *JWTMiddleware {
	return &JWTMiddleware{
		jwtManager: jwtManager,
	}
}

func (m *JWTMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	const op = "middlewares.jwt.Handle"
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uinfo, err := m.jwtManager.GetUserInfo(r)
		if err != nil {
			http.Error(w, fmt.Sprintf("%s: get user info: %s", op, err.Error()), http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserCtxKey, uinfo)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
