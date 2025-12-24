package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/pubflow/flowfull-go-starter/internal/lib/cache"
	"github.com/pubflow/flowfull-go-starter/internal/lib/database"
)

// HealthRoutes sets up health check routes
func HealthRoutes(app *fiber.App, db *database.Connection, hybridCache *cache.HybridCache, redisClient *redis.Client, logger *zap.Logger) {
	health := app.Group("/health")

	// Basic health check
	health.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"service": "flowfull-go-starter",
		})
	})

	// Database health check
	health.Get("/db", func(c *fiber.Ctx) error {
		sqlDB, err := db.DB.DB()
		if err != nil {
			logger.Error("failed to get sql.DB", zap.Error(err))
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"status": "error",
				"error":  "database connection failed",
			})
		}

		if err := sqlDB.Ping(); err != nil {
			logger.Error("database ping failed", zap.Error(err))
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"status": "error",
				"error":  "database ping failed",
			})
		}

		return c.JSON(fiber.Map{
			"status":   "ok",
			"database": "connected",
		})
	})

	// Cache health check
	health.Get("/cache", func(c *fiber.Ctx) error {
		if redisClient == nil {
			return c.JSON(fiber.Map{
				"status": "ok",
				"cache":  "disabled",
			})
		}

		if err := redisClient.Ping(c.Context()).Err(); err != nil {
			logger.Error("redis ping failed", zap.Error(err))
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"status": "error",
				"error":  "redis ping failed",
			})
		}

		metrics := hybridCache.GetMetrics()
		return c.JSON(fiber.Map{
			"status": "ok",
			"cache":  "connected",
			"metrics": fiber.Map{
				"ristretto_hits":   metrics.RistrettoHits,
				"ristretto_misses": metrics.RistrettoMisses,
				"redis_hits":       metrics.RedisHits,
				"redis_misses":     metrics.RedisMisses,
			},
		})
	})

	// Complete health check
	health.Get("/all", func(c *fiber.Ctx) error {
		status := fiber.Map{
			"status":  "ok",
			"service": "flowfull-go-starter",
		}

		// Check database
		sqlDB, err := db.DB.DB()
		if err != nil || sqlDB.Ping() != nil {
			status["database"] = "error"
			status["status"] = "degraded"
		} else {
			status["database"] = "ok"
		}

		// Check cache
		if redisClient != nil {
			if err := redisClient.Ping(c.Context()).Err(); err != nil {
				status["cache"] = "error"
				status["status"] = "degraded"
			} else {
				status["cache"] = "ok"
			}
		} else {
			status["cache"] = "disabled"
		}

		if status["status"] == "degraded" {
			return c.Status(fiber.StatusServiceUnavailable).JSON(status)
		}

		return c.JSON(status)
	})
}

