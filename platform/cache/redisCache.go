package cache

import (
	"context"
	"strings"
	"time"

	"platform/logger"

	fb "github.com/Eugene-Usachev/fastbytes"
	"github.com/caarlos0/env/v11"
	"github.com/goccy/go-json"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/rueidis"
)

type RedisCache struct {
	client      rueidis.Client
	cacheHits   *prometheus.CounterVec
	cacheMisses *prometheus.CounterVec
}

const (
	cacheDurationSeconds        = 300
	negativeCaseDurationSeconds = 300
)

type redisConfig struct {
	Addrs    string `env:"REDIS_ADDRS, required, notEmpty"`
	Password string `env:"REDIS_PASSWORD, required, notEmpty"`
}

func MustCreateRedisCache() *RedisCache {
	cfg, err := env.ParseAs[redisConfig]()
	if err != nil {
		logger.Fatal(err.Error())

		return nil
	}

	var addrsArr []string

	if err = json.Unmarshal([]byte(cfg.Addrs), &addrsArr); err != nil {
		logger.Fatalf("error occurred when unmarshalling redis Addrs: %v", err)

		return nil
	}

	if len(addrsArr) == 0 {
		logger.Fatal("error occurred when creating a redis client: empty address")

		return nil
	}

	client, err := rueidis.NewClient(rueidis.ClientOption{
		InitAddress:   addrsArr,
		Password:      cfg.Password,
		SelectDB:      0,
		MaxFlushDelay: 100 * time.Microsecond,
	})
	if err != nil {
		logger.Fatalf("error occurred when creating a redis client: %s", err.Error())
	}

	cacheHits := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_hits_total",
			Help: "Number of cache hits by metric name",
		},
		[]string{"table"},
	)
	cacheMisses := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_misses_total",
			Help: "Number of cache misses by metric name",
		},
		[]string{"table"},
	)

	prometheus.MustRegister(cacheHits, cacheMisses)

	return &RedisCache{
		client:      client,
		cacheHits:   cacheHits,
		cacheMisses: cacheMisses,
	}
}

var _ Cache = (*RedisCache)(nil)

func redisKey(table string, key string) string {
	builder := strings.Builder{}

	builder.Grow(len(table) + len(key) + 1)
	builder.WriteString(table)
	builder.WriteByte(':')
	builder.WriteString(key)

	return builder.String()
}

func (cache *RedisCache) IsNegativeCase(ctx context.Context, table string, key string) bool {
	realKey := redisKey(table, key)

	res, err := cache.client.Do(ctx, cache.client.B().Exists().Key(realKey).Build()).AsBool()
	if err != nil {
		if !rueidis.IsRedisNil(err) {
			logger.Errorf(
				"[Redis] error occurred when getting string by key: %s, error: %s",
				key, err.Error(),
			)
		} else {
			cache.cacheMisses.WithLabelValues(table).Inc()
		}

		return false
	}

	cache.cacheHits.WithLabelValues(table).Inc()

	return res
}

func (cache *RedisCache) GetString(
	ctx context.Context,
	table string,
	key string,
) (string, bool) {
	realKey := redisKey(table, key)

	res, err := cache.client.Do(ctx, cache.client.B().Get().Key(realKey).Build()).ToString()
	if err != nil {
		if !rueidis.IsRedisNil(err) {
			logger.Errorf(
				"[Redis] error occurred when getting string by key: %s, error: %s",
				key,
				err.Error(),
			)
		} else {
			cache.cacheMisses.WithLabelValues(table).Inc()
		}

		return "", false
	}

	cache.cacheHits.WithLabelValues(table).Inc()

	return res, true
}

func (cache *RedisCache) GetBytes(
	ctx context.Context,
	table string,
	key string,
) ([]byte, bool) {
	realKey := redisKey(table, key)

	res, err := cache.client.Do(ctx, cache.client.B().Get().Key(realKey).Build()).AsBytes()
	if err != nil {
		if !rueidis.IsRedisNil(err) {
			logger.Errorf(
				"[Redis] error occurred when getting string by key: %s, error: %s",
				key,
				err.Error(),
			)
		} else {
			cache.cacheMisses.WithLabelValues(table).Inc()
		}

		return nil, false
	}

	cache.cacheHits.WithLabelValues(table).Inc()

	return res, true
}

func (cache *RedisCache) SetString(ctx context.Context, table string, key string, value string) {
	realKey := redisKey(table, key)

	if err := cache.client.Do(
		ctx,
		cache.client.B().Set().Key(realKey).Value(value).ExSeconds(cacheDurationSeconds).Build(),
	).Error(); err != nil {
		logger.Errorf(
			"[Redis] error occurred when setting string by key: %s, error: %s",
			key,
			err.Error(),
		)
	}
}

func (cache *RedisCache) SetBytes(ctx context.Context, table string, key string, value []byte) {
	realKey := redisKey(table, key)

	if err := cache.client.Do(
		ctx,
		cache.client.B().Set().Key(realKey).Value(fb.B2S(value)).ExSeconds(cacheDurationSeconds).Build(),
	).Error(); err != nil {
		logger.Errorf(
			"[Redis] error occurred when setting bytes by key: %s, error: %s",
			key, err.Error(),
		)
	}
}

func (cache *RedisCache) SetNegativeCase(ctx context.Context, table string, key string) {
	realKey := redisKey(table, key)

	if err := cache.client.Do(
		ctx,
		cache.client.B().Set().Key(realKey).Value("").ExSeconds(negativeCaseDurationSeconds).Build(),
	).Error(); err != nil {
		logger.Errorf(
			"[Redis] error occurred when setting negative case by key: %s, error: %s",
			key, err.Error(),
		)
	}
}

func (cache *RedisCache) Delete(ctx context.Context, table string, key string) error {
	realKey := redisKey(table, key)

	return cache.client.Do(ctx, cache.client.B().Del().Key(realKey).Build()).Error()
}
