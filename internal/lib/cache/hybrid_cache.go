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

// HybridCache implements 3-tier caching: Ristretto → Redis → Database
type HybridCache struct {
	ristretto *ristretto.Cache
	redis     *redis.Client
	logger    *zap.Logger
	metrics   *CacheMetrics
	enabled   bool
}

// NewHybridCache creates a new HybridCache instance
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

	// Configure Ristretto
	ristrettoCache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: maxSize * 10,
		MaxCost:     maxSize,
		BufferItems: 64,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create ristretto cache: %w", err)
	}

	logger.Info("hybrid cache initialized",
		zap.Int64("max_size", maxSize),
		zap.Bool("redis_enabled", redisClient != nil),
	)

	return &HybridCache{
		ristretto: ristrettoCache,
		redis:     redisClient,
		logger:    logger,
		metrics:   &CacheMetrics{},
		enabled:   true,
	}, nil
}

// Get retrieves a value from cache (Ristretto → Redis → nil)
func (hc *HybridCache) Get(ctx context.Context, key string) (interface{}, bool) {
	if !hc.enabled {
		return nil, false
	}

	// Try Ristretto first
	if value, found := hc.ristretto.Get(key); found {
		hc.metrics.RistrettoHits++
		hc.logger.Debug("ristretto cache hit", zap.String("key", key))
		return value, true
	}
	hc.metrics.RistrettoMisses++

	// Try Redis
	if hc.redis != nil {
		val, err := hc.redis.Get(ctx, key).Result()
		if err == nil {
			hc.metrics.RedisHits++
			hc.logger.Debug("redis cache hit", zap.String("key", key))

			// Backfill Ristretto
			var data interface{}
			if err := json.Unmarshal([]byte(val), &data); err == nil {
				hc.ristretto.Set(key, data, 1)
			}

			return data, true
		}
		hc.metrics.RedisMisses++
	}

	return nil, false
}

// Set stores a value in all cache tiers
func (hc *HybridCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if !hc.enabled {
		return nil
	}

	// Set in Ristretto
	hc.ristretto.SetWithTTL(key, value, 1, ttl)

	// Set in Redis
	if hc.redis != nil {
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
	}

	hc.logger.Debug("cache set", zap.String("key", key), zap.Duration("ttl", ttl))
	return nil
}

// Delete removes a value from all cache tiers
func (hc *HybridCache) Delete(ctx context.Context, key string) error {
	if !hc.enabled {
		return nil
	}

	// Delete from Ristretto
	hc.ristretto.Del(key)

	// Delete from Redis
	if hc.redis != nil {
		if err := hc.redis.Del(ctx, key).Err(); err != nil {
			hc.logger.Error("failed to delete from redis",
				zap.Error(err),
				zap.String("key", key),
			)
			return err
		}
	}

	hc.logger.Debug("cache deleted", zap.String("key", key))
	return nil
}

// GetMetrics returns cache metrics
func (hc *HybridCache) GetMetrics() *CacheMetrics {
	return hc.metrics
}

// Close closes the cache
func (hc *HybridCache) Close() {
	if hc.ristretto != nil {
		hc.ristretto.Close()
	}
}

