// Package middlewares ratelimiter.go implements
// token bucket ratelimiting via redis
package middlewares

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"gateway/internal/security"
	"gateway/internal/utils"

	"github.com/go-redis/redis/v8"
)

type RateLimiter struct {
	capacity   int
	fillRate   float64
	ctxTimeout time.Duration
	rdb        *redis.Client
}

func NewRateLimiter(ctxTimeout time.Duration) (Middleware, error) {
	const op = "middlewares.NewRateLimiter"

	pingTimeout := time.Duration(utils.GetEnvInt("REDIS_PING_TIMEOUT", 10)) * time.Second

	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_RL_ADDR"),
		Password: os.Getenv("REDIS_RL_PASSWORD"),
		DB:       0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), pingTimeout)
	defer cancel()

	if _, err := rdb.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("%s: ping: %w", op, err)
	}

	limit := utils.GetEnvInt("RATELIMIT_CAPACITY", 50)
	fillRate := float64(utils.GetEnvInt("RATELIMIT_FILL_RATE", 3))

	return &RateLimiter{
		capacity:   limit,
		fillRate:   fillRate,
		ctxTimeout: ctxTimeout,
		rdb:        rdb,
	}, nil
}

var ratelimitScript = redis.NewScript(`
local key = KEYS[1]
local capacity = tonumber(ARGV[1])
local fill_rate = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
local requested = 1

local data = redis.call('HMGET', key, 'tokens', 'last_updated')
local tokens = tonumber(data[1])
local last_updated = tonumber(data[2])

if tokens == nil then
	tokens = capacity
	last_updated = now
else
	local delta = math.max(0, (now - last_updated) / 1000)
	local tokens_to_add = delta * fill_rate
	tokens = math.min(capacity, tokens + tokens_to_add)
	last_updated = now
end

if tokens >= requested then
	tokens = tokens - requested
	redis.call('HMSET', key, 'tokens', tokens, 'last_updated', last_updated)
	redis.call('EXPIRE', key, math.ceil(capacity / fill_rate) * 2)
	return 1
else
	redis.call('HMSET', key, 'tokens', tokens, 'last_updated', last_updated)
	return 0
end
`)

func (rl *RateLimiter) Handle(next http.HandlerFunc) http.HandlerFunc {
	const op = "middlewares.ratelimiter.Handle"

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := rl.generateKey(r)

		ctx, cancel := context.WithTimeout(r.Context(), rl.ctxTimeout)
		defer cancel()

		now := time.Now().UnixMilli()
		res, err := ratelimitScript.Run(ctx, rl.rdb, []string{key}, rl.capacity, rl.fillRate, now).Int()
		if err != nil {
			http.Error(w, fmt.Sprintf("%s: run script: %s", op, err.Error()), http.StatusInternalServerError)
			return
		}

		if res == 0 {
			http.Error(w, fmt.Sprintf("%s: ratelimiter: too many requests", op), http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (rl *RateLimiter) generateKey(r *http.Request) string {
	uinfo, ok := r.Context().Value(UserCtxKey).(security.UserInfo)
	if ok && uinfo.Nickname != "" {
		return fmt.Sprintf("rl:usr:%s", uinfo.Nickname)
	}

	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err == nil {
			ip = host
		} else {
			ip = r.RemoteAddr
		}
	}

	raw := fmt.Sprintf("%s|%s", ip, r.UserAgent())
	hash := sha256.Sum256([]byte(raw))

	return fmt.Sprintf("rl:anon:%s", hex.EncodeToString(hash[:]))
}
