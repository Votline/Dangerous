// Package rdb provides redis database interface and
// it implementation by MarksCache for caching
package rdb

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"
	"unsafe"

	"mrksrv/internal/utils"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

type RDB interface {
	GetOrIncr(ctx context.Context, key string) ([]byte, error)
	SetIfHot(ctx context.Context, key string, val []byte) error
	DelIfExist(ctx context.Context, key string) error
}

type MarksCache struct {
	cntLimit   int
	expCounter int
	expCache   int
	rdb        *redis.Client
	log        *zap.Logger
}

func NewMarksCache(log *zap.Logger) (RDB, error) {
	const op = "rdb.NewMarksCache"

	pingTimeout := time.Duration(utils.GetEnvInt("REDIS_PING_TIMEOUT", 10)) * time.Second
	expCounter := time.Duration(utils.GetEnvInt("REDIS_MC_EXP_CNT", 10)) * time.Minute
	expCache := time.Duration(utils.GetEnvInt("REDIS_MC_EXP_CACHE", 15)) * time.Minute
	cntLimit := utils.GetEnvInt("REDIS_MC_CNT_LIMIT", 30)

	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_MC_ADDR"),
		Password: os.Getenv("REDIS_MC_PASSWORD"),
		DB:       0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), pingTimeout)
	defer cancel()

	if _, err := rdb.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("%s: ping: %w", op, err)
	}

	return &MarksCache{
		log:        log,
		cntLimit:   cntLimit,
		expCounter: int(expCounter.Seconds()),
		expCache:   int(expCache.Seconds()),
		rdb:        rdb,
	}, nil
}

var getOrIncrScript = redis.NewScript(`
local cntKey = KEYS[1]
local dataKey = KEYS[2]
local limit = tonumber(ARGV[1])
local expCounterSec = tonumber(ARGV[2])

local count = redis.call('INCR', cntKey)
if count == 1 then
    redis.call('EXPIRE', cntKey, expCounterSec)
end

if count >= limit then
    return redis.call('GET', dataKey)
end

return nil
`)

func (rdb *MarksCache) GetOrIncr(ctx context.Context, key string) ([]byte, error) {
	const op = "rdb.MarksCache.GetOrIncr"

	cntKey := "marks:cnt:" + key
	dataKey := "marks:data:" + key

	rdb.log.Debug("GetOrIncr request",
		zap.String("op", op),
		zap.String("cntkey", cntKey),
		zap.String("datakey", dataKey))

	res, err := getOrIncrScript.Run(ctx, rdb.rdb, []string{cntKey, dataKey}, rdb.cntLimit, rdb.expCounter).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, fmt.Errorf("%s: redis get script: %w", op, err)
	}

	str, ok := res.(string)
	if !ok {
		return nil, fmt.Errorf("%s: type assert: failed", op)
	}

	if str == "" {
		return nil, nil
	}

	resSlc := unsafe.Slice(unsafe.StringData(str), len(str))

	rdb.log.Debug("get result",
		zap.String("op", op),
		zap.Int("len", len(resSlc)))

	return resSlc, nil
}

var setIfHotScript = redis.NewScript(`
local cntKey = KEYS[1]
local dataKey = KEYS[2]
local limit = tonumber(ARGV[1])
local val = ARGV[2]
local expCacheSec = tonumber(ARGV[3])

local count = tonumber(redis.call('GET', cntKey) or "0")

if count >= limit then
    redis.call('SET', dataKey, val, 'EX', expCacheSec)
    return 1
end

return 0
`)

func (rdb *MarksCache) SetIfHot(ctx context.Context, key string, val []byte) error {
	const op = "rdb.MarksCache.SetIfHot"

	cntKey := "marks:cnt:" + key
	dataKey := "marks:data:" + key

	rdb.log.Debug("GetOrIncr request",
		zap.String("op", op),
		zap.String("cntkey", cntKey),
		zap.String("datakey", dataKey))

	if _, err := setIfHotScript.Run(ctx, rdb.rdb, []string{cntKey, dataKey}, rdb.cntLimit, val, rdb.expCache).Result(); err != nil {
		if errors.Is(err, redis.Nil) {
			return nil
		}
		return fmt.Errorf("%s: redis set script: %w", op, err)
	}

	rdb.log.Debug("successfully setted",
		zap.String("op", op))

	return nil
}

func (rdb *MarksCache) DelIfExist(ctx context.Context, key string) error {
	const op = "rdb.MarksCache.DelIfExist"

	dataKey := "marks:data:" + key

	if err := rdb.rdb.Del(ctx, dataKey).Err(); err != nil {
		return fmt.Errorf("%s: delete data: %w", op, err)
	}

	return nil
}
