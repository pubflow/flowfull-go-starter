package routes

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/pubflow/flowfull-go-starter/internal/config"
	"github.com/pubflow/flowfull-go-starter/internal/lib/auth"
	"github.com/pubflow/flowfull-go-starter/internal/lib/database"
)

// APIRoutes sets up API routes
func APIRoutes(app *fiber.App, db *database.Connection, authMiddleware *auth.AuthMiddleware, logger *zap.Logger, cfg *config.Config) {
	api := app.Group("/api")

	// Public route (no authentication required)
	api.Get("/public", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "This is a public endpoint",
			"data":    "Anyone can access this",
		})
	})

	// Test endpoint - shows current auth configuration
	api.Get("/test/config", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"validation_mode": cfg.AuthValidationMode,
			"flowless_url":    cfg.FlowlessAPIURL,
			"cache_enabled":   cfg.CacheEnabled,
			"environment":     cfg.Environment,
			"message":         "Current authentication configuration",
			"help": fiber.Map{
				"modes": fiber.Map{
					"DEVELOPMENT": "No validation - accepts any session ID",
					"PERMISSIVE":  "Validates format only, no bridge check",
					"STANDARD":    "Full validation with Flowless (default)",
					"STRICT":      "Full validation + IP/UA checks",
				},
				"to_test_without_flowless": "Set AUTH_VALIDATION_MODE=DEVELOPMENT in .env",
			},
		})
	})

	// Protected route (authentication required)
	api.Get("/protected", authMiddleware.RequireAuth(), func(c *fiber.Ctx) error {
		userID := c.Locals("user_id").(string)
		email := c.Locals("email").(string)

		return c.JSON(fiber.Map{
			"message": "This is a protected endpoint",
			"user": fiber.Map{
				"id":    userID,
				"email": email,
			},
		})
	})

	// Optional auth route (authentication optional)
	api.Get("/optional", authMiddleware.OptionalAuth(), func(c *fiber.Ctx) error {
		userID := c.Locals("user_id")
		if userID != nil {
			return c.JSON(fiber.Map{
				"message":        "Authenticated user",
				"user_id":        userID,
				"authenticated":  true,
			})
		}

		return c.JSON(fiber.Map{
			"message":       "Anonymous user",
			"authenticated": false,
		})
	})

	// User profile route
	api.Get("/profile", authMiddleware.RequireAuth(), func(c *fiber.Ctx) error {
		sessionData := c.Locals("session_data").(*auth.SessionData)

		return c.JSON(fiber.Map{
			"user": fiber.Map{
				"id":              sessionData.UserID,
				"email":           sessionData.Email,
				"name":            sessionData.Name,
				"user_type":       sessionData.UserType,
				"organization_id": sessionData.OrganizationID,
				"permissions":     sessionData.Permissions,
			},
		})
	})

	// ========================================
	// TASKS CRUD ROUTES - MOCK ENDPOINTS
	// ========================================
	// These are example endpoints to demonstrate API structure
	// NO DATABASE OPERATIONS - All responses are mock data
	// Replace with real database logic when needed
	// ========================================
	tasks := api.Group("/tasks", authMiddleware.RequireAuth())

	// List tasks (MOCK)
	tasks.Get("/", func(c *fiber.Ctx) error {
		userID := c.Locals("user_id").(string)

		// Mock data - no database query
		mockTasks := []fiber.Map{
			{
				"id":          "1",
				"user_id":     userID,
				"title":       "Example Task 1",
				"description": "This is a mock task",
				"completed":   false,
				"created_at":  "2024-01-01T00:00:00Z",
			},
			{
				"id":          "2",
				"user_id":     userID,
				"title":       "Example Task 2",
				"description": "Another mock task",
				"completed":   true,
				"created_at":  "2024-01-02T00:00:00Z",
			},
		}

		return c.JSON(fiber.Map{
			"tasks":   mockTasks,
			"count":   len(mockTasks),
			"message": "Mock data - not stored in database",
		})
	})

	// Create task (MOCK)
	tasks.Post("/", func(c *fiber.Ctx) error {
		userID := c.Locals("user_id").(string)

		var requestBody fiber.Map
		if err := c.BodyParser(&requestBody); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid request body",
			})
		}

		// Mock response - not saved to database
		mockTask := fiber.Map{
			"id":          "mock-" + userID + "-123",
			"user_id":     userID,
			"title":       requestBody["title"],
			"description": requestBody["description"],
			"completed":   false,
			"created_at":  "2024-01-01T00:00:00Z",
		}

		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"message": "Mock task created (not saved to database)",
			"task":    mockTask,
		})
	})

	// Get task by ID (MOCK)
	tasks.Get("/:id", func(c *fiber.Ctx) error {
		userID := c.Locals("user_id").(string)
		taskID := c.Params("id")

		// Mock response
		mockTask := fiber.Map{
			"id":          taskID,
			"user_id":     userID,
			"title":       "Mock Task " + taskID,
			"description": "This is a mock task retrieved by ID",
			"completed":   false,
			"created_at":  "2024-01-01T00:00:00Z",
		}

		return c.JSON(fiber.Map{
			"task":    mockTask,
			"message": "Mock data - not from database",
		})
	})

	// Update task (MOCK)
	tasks.Put("/:id", func(c *fiber.Ctx) error {
		userID := c.Locals("user_id").(string)
		taskID := c.Params("id")

		var requestBody fiber.Map
		if err := c.BodyParser(&requestBody); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid request body",
			})
		}

		// Mock response - not saved to database
		mockTask := fiber.Map{
			"id":          taskID,
			"user_id":     userID,
			"title":       requestBody["title"],
			"description": requestBody["description"],
			"completed":   requestBody["completed"],
			"updated_at":  "2024-01-01T00:00:00Z",
		}

		return c.JSON(fiber.Map{
			"message": "Mock task updated (not saved to database)",
			"task":    mockTask,
		})
	})

	// Delete task (MOCK)
	tasks.Delete("/:id", func(c *fiber.Ctx) error {
		taskID := c.Params("id")

		// Mock response - nothing deleted from database
		return c.JSON(fiber.Map{
			"message":  "Mock task deleted (not removed from database)",
			"task_id":  taskID,
			"deleted":  true,
			"is_mock":  true,
		})
	})
}

