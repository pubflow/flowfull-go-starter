package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// CacheMetrics tracks cache performance
type CacheMetrics struct {
	RistrettoHits   int64
	RistrettoMisses int64
	RedisHits       int64
	RedisMisses     int64
	DatabaseHits    int64
}

// HybridCache uses exactly one backend: Redis when connected, otherwise local Ristretto.
// Database remains the source of truth outside this cache.
type HybridCache struct {
	ristretto *ristretto.Cache
	redis     *redis.Client
	logger    *zap.Logger
	metrics   *CacheMetrics
	enabled   bool
}

// NewHybridCache creates a new HybridCache instance.
// If redisClient is non-nil it is the only store; Ristretto is not allocated.
func NewHybridCache(
	maxSize int64,
	redisClient *redis.Client,
	logger *zap.Logger,
	enabled bool,
) (*HybridCache, error) {
	if !enabled {
		logger.Info("cache disabled")
		return &HybridCache{
			enabled: false,
			logger:  logger,
			metrics: &CacheMetrics{},
		}, nil
	}

	hc := &HybridCache{
		redis:   redisClient,
		logger:  logger,
		metrics: &CacheMetrics{},
		enabled: true,
	}

	if redisClient != nil {
		logger.Info("hybrid cache initialized",
			zap.String("backend", "redis"),
		)
		return hc, nil
	}

	ristrettoCache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: maxSize * 10,
		MaxCost:     maxSize,
		BufferItems: 64,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create ristretto cache: %w", err)
	}

	hc.ristretto = ristrettoCache
	logger.Info("hybrid cache initialized",
		zap.String("backend", "ristretto"),
		zap.Int64("max_size", maxSize),
	)
	return hc, nil
}

func (hc *HybridCache) useRedis() bool {
	return hc.redis != nil
}

// RedisEnabled reports whether Redis is the exclusive cache backend.
func (hc *HybridCache) RedisEnabled() bool {
	return hc.enabled && hc.useRedis()
}

// Get retrieves a value from the active cache backend.
func (hc *HybridCache) Get(ctx context.Context, key string) (interface{}, bool) {
	if !hc.enabled {
		return nil, false
	}

	if hc.useRedis() {
		val, err := hc.redis.Get(ctx, key).Result()
		if err != nil {
			hc.metrics.RedisMisses++
			return nil, false
		}
		hc.metrics.RedisHits++
		hc.logger.Debug("redis cache hit", zap.String("key", key))

		var data interface{}
		if err := json.Unmarshal([]byte(val), &data); err != nil {
			return val, true
		}
		return data, true
	}

	if hc.ristretto == nil {
		return nil, false
	}

	if value, found := hc.ristretto.Get(key); found {
		hc.metrics.RistrettoHits++
		hc.logger.Debug("ristretto cache hit", zap.String("key", key))
		return value, true
	}
	hc.metrics.RistrettoMisses++
	return nil, false
}

// Set stores a value in the active cache backend only.
func (hc *HybridCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if !hc.enabled {
		return nil
	}

	if hc.useRedis() {
		data, err := json.Marshal(value)
		if err != nil {
			hc.logger.Error("failed to marshal value for redis",
				zap.Error(err),
				zap.String("key", key),
			)
			return err
		}

		if err := hc.redis.Set(ctx, key, data, ttl).Err(); err != nil {
			hc.logger.Error("failed to set redis cache",
				zap.Error(err),
				zap.String("key", key),
			)
			return err
		}
		hc.logger.Debug("cache set", zap.String("key", key), zap.Duration("ttl", ttl), zap.String("backend", "redis"))
		return nil
	}

	if hc.ristretto != nil {
		hc.ristretto.SetWithTTL(key, value, 1, ttl)
		hc.ristretto.Wait()
	}

	hc.logger.Debug("cache set", zap.String("key", key), zap.Duration("ttl", ttl), zap.String("backend", "ristretto"))
	return nil
}

// Delete removes a value from the active cache backend only.
func (hc *HybridCache) Delete(ctx context.Context, key string) error {
	if !hc.enabled {
		return nil
	}

	if hc.useRedis() {
		if err := hc.redis.Del(ctx, key).Err(); err != nil {
			hc.logger.Error("failed to delete from redis",
				zap.Error(err),
				zap.String("key", key),
			)
			return err
		}
		hc.logger.Debug("cache deleted", zap.String("key", key), zap.String("backend", "redis"))
		return nil
	}

	if hc.ristretto != nil {
		hc.ristretto.Del(key)
	}

	hc.logger.Debug("cache deleted", zap.String("key", key), zap.String("backend", "ristretto"))
	return nil
}

// GetMetrics returns cache metrics
func (hc *HybridCache) GetMetrics() *CacheMetrics {
	return hc.metrics
}

// Close closes the local cache if it was created.
func (hc *HybridCache) Close() {
	if hc.ristretto != nil {
		hc.ristretto.Close()
	}
}
