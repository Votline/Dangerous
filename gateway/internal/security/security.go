// Package security implements help-functions for
// manage jwt tokens
package security

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"gateway/internal/utils"

	"github.com/golang-jwt/jwt/v5"
)

type JWTSecurity interface {
	Init() error
	GetUserInfo(r *http.Request) (UserInfo, error)
	GenerateToken(nickname string) (string, error)
	ExtractClaims(tokenStr string) (UserInfo, error)
	ExtractUnverifiedClaims(tokenStr string) (UserInfo, error)
}

type JWTManager struct {
	secret []byte
	exp    time.Duration
}

func (j *JWTManager) Init() error {
	const op = "security.JWTManager.Init"

	j.secret = []byte(os.Getenv("JWT_SECRET"))
	j.exp = time.Duration(utils.GetEnvInt("JWT_EXP", 15)) * time.Minute

	if len(j.secret) == 0 {
		return fmt.Errorf("%s: get jwt secret: no jwt secret", op)
	}

	return nil
}

type UserInfo struct {
	Nickname string `json:"nickname"`
	token    *jwt.Token
	jwt.RegisteredClaims
}

func (j *JWTManager) GenerateToken(nickname string) (string, error) {
	const op = "security.JWTManager.GenerateToken"

	claims := UserInfo{
		Nickname: nickname,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.exp)),
		},
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(j.secret)
	if err != nil {
		return "", fmt.Errorf("%s: create token: %w", op, err)
	}

	return token, nil
}

func (j *JWTManager) ExtractClaims(tokenStr string) (UserInfo, error) {
	const op = "security.JWTManager.ExtractClaims"

	claims, err := j.ExtractUnverifiedClaims(tokenStr)
	if err != nil {
		return UserInfo{}, fmt.Errorf("%s: extract: %w", op, err)
	}

	if !claims.token.Valid {
		return UserInfo{}, fmt.Errorf("%s: token validation: token is invalid", op)
	}

	return claims, nil
}

func (j *JWTManager) ExtractUnverifiedClaims(tokenStr string) (UserInfo, error) {
	const op = "security.JWTManager.ExtractUnverifiedClaims"

	claims := UserInfo{}
	parser := jwt.NewParser(jwt.WithoutClaimsValidation())
	token, err := parser.ParseWithClaims(tokenStr, &claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("%s: unexpected signing method: %v", op, token.Header["alg"])
		}
		return j.secret, nil
	})
	if err != nil {
		return UserInfo{}, fmt.Errorf("%s: parse with claims: %w", op, err)
	}

	claims.token = token

	return claims, nil
}

func (j *JWTManager) GetUserInfo(r *http.Request) (UserInfo, error) {
	const op = "security.GetUserInfo"

	cookie, err := r.Cookie("token")
	if err != nil {
		return UserInfo{}, fmt.Errorf("%s: get cookie: %w", op, err)
	}

	uinfo, err := j.ExtractClaims(cookie.Value)
	if err != nil {
		return UserInfo{}, fmt.Errorf("%s: extract claims: %w", op, err)
	}

	return uinfo, nil
}
