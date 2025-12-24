package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/pubflow/flowfull-go-starter/internal/config"
	"github.com/pubflow/flowfull-go-starter/internal/lib/auth"
	"github.com/pubflow/flowfull-go-starter/internal/lib/cache"
	"github.com/pubflow/flowfull-go-starter/internal/lib/database"
	"github.com/pubflow/flowfull-go-starter/internal/lib/utils"
	"github.com/pubflow/flowfull-go-starter/internal/routes"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger
	zapLogger, err := utils.NewLogger(cfg.LogLevel, cfg.LogFormat, cfg.IsDevelopment())
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer zapLogger.Sync()

	zapLogger.Info("starting flowfull-go-starter",
		zap.String("environment", cfg.Environment),
		zap.Int("port", cfg.Port),
	)

	// Initialize database
	db, err := database.NewConnection(cfg, zapLogger)
	if err != nil {
		zapLogger.Fatal("failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// Note: Auto-migrations removed - manage your database schema manually
	// or use a migration tool like golang-migrate, goose, or atlas

	// Initialize Redis (optional)
	var redisClient *redis.Client
	if cfg.RedisURL != "" {
		opt, err := redis.ParseURL(cfg.RedisURL)
		if err != nil {
			zapLogger.Warn("failed to parse redis URL", zap.Error(err))
		} else {
			redisClient = redis.NewClient(opt)
			if err := redisClient.Ping(context.Background()).Err(); err != nil {
				zapLogger.Warn("failed to connect to redis", zap.Error(err))
				redisClient = nil
			} else {
				zapLogger.Info("redis connected")
			}
		}
	}

	// Initialize HybridCache
	hybridCache, err := cache.NewHybridCache(
		cfg.CacheMaxSize,
		redisClient,
		zapLogger,
		cfg.CacheEnabled,
	)
	if err != nil {
		zapLogger.Fatal("failed to initialize cache", zap.Error(err))
	}
	defer hybridCache.Close()

	// Initialize Bridge Validator
	bridgeValidator := auth.NewBridgeValidator(cfg, zapLogger)

	// Initialize Auth Middleware
	authMiddleware := auth.NewAuthMiddleware(bridgeValidator, hybridCache, cfg, zapLogger)

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName:      "Flowfull Go Starter",
		ErrorHandler: customErrorHandler,
	})

	// Middleware
	app.Use(recover.New())

	if cfg.DevLogRequests {
		app.Use(logger.New())
	}

	// CORS
	app.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORSOrigins,
		AllowMethods:     cfg.CORSMethods,
		AllowHeaders:     cfg.CORSHeaders,
		AllowCredentials: cfg.CORSCredentials,
		MaxAge:           cfg.CORSMaxAge,
	}))

	// Routes
	routes.HealthRoutes(app, db, hybridCache, redisClient, zapLogger)
	routes.APIRoutes(app, db, authMiddleware, zapLogger, cfg)

	// Root route
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"service":     "flowfull-go-starter",
			"version":     "1.0.0",
			"environment": cfg.Environment,
			"status":      "running",
		})
	})

	// Start server in goroutine
	go func() {
		addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
		zapLogger.Info("server starting", zap.String("address", addr))
		if err := app.Listen(addr); err != nil {
			zapLogger.Fatal("failed to start server", zap.Error(err))
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	zapLogger.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		zapLogger.Error("server forced to shutdown", zap.Error(err))
	}

	zapLogger.Info("server stopped")
}

// customErrorHandler handles errors globally
func customErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError

	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}

	return c.Status(code).JSON(fiber.Map{
		"error":   err.Error(),
		"code":    code,
		"path":    c.Path(),
		"method":  c.Method(),
	})
}

