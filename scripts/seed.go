package main

import (
	"fmt"
	"log"
	"time"

	"github.com/pubflow/flowfull-go-starter/internal/config"
	"github.com/pubflow/flowfull-go-starter/internal/lib/database"
	"github.com/pubflow/flowfull-go-starter/internal/lib/utils"
	"github.com/pubflow/flowfull-go-starter/internal/models"
)

func main() {
	fmt.Println("🌱 Seeding database...")

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger
	logger, err := utils.NewLogger(cfg.LogLevel, cfg.LogFormat, cfg.IsDevelopment())
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	// Connect to database
	db, err := database.NewConnection(cfg, logger)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := db.AutoMigrate(&models.Task{}); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Seed data
	seedTasks(db)

	fmt.Println("✅ Database seeded successfully!")
}

func seedTasks(db *database.Connection) {
	// Sample user IDs (replace with actual user IDs from your Flowless instance)
	userIDs := []string{
		"user_123",
		"user_456",
		"user_789",
	}

	tasks := []models.Task{
		// User 1 tasks
		{
			UserID:      userIDs[0],
			Title:       "Complete project documentation",
			Description: "Write comprehensive documentation for the new feature",
			Status:      "in_progress",
			Priority:    "high",
			DueDate:     timePtr(time.Now().AddDate(0, 0, 7)),
		},
		{
			UserID:      userIDs[0],
			Title:       "Review pull requests",
			Description: "Review and merge pending pull requests",
			Status:      "pending",
			Priority:    "medium",
			DueDate:     timePtr(time.Now().AddDate(0, 0, 3)),
		},
		{
			UserID:      userIDs[0],
			Title:       "Update dependencies",
			Description: "Update all project dependencies to latest versions",
			Status:      "pending",
			Priority:    "low",
			DueDate:     timePtr(time.Now().AddDate(0, 0, 14)),
		},

		// User 2 tasks
		{
			UserID:      userIDs[1],
			Title:       "Design new landing page",
			Description: "Create mockups for the new landing page",
			Status:      "in_progress",
			Priority:    "high",
			DueDate:     timePtr(time.Now().AddDate(0, 0, 5)),
		},
		{
			UserID:      userIDs[1],
			Title:       "Conduct user research",
			Description: "Interview 10 users about their experience",
			Status:      "pending",
			Priority:    "medium",
			DueDate:     timePtr(time.Now().AddDate(0, 0, 10)),
		},

		// User 3 tasks
		{
			UserID:      userIDs[2],
			Title:       "Fix critical bug in production",
			Description: "Investigate and fix the authentication issue",
			Status:      "in_progress",
			Priority:    "high",
			DueDate:     timePtr(time.Now().AddDate(0, 0, 1)),
		},
		{
			UserID:      userIDs[2],
			Title:       "Optimize database queries",
			Description: "Improve performance of slow queries",
			Status:      "pending",
			Priority:    "medium",
			DueDate:     timePtr(time.Now().AddDate(0, 0, 7)),
		},
		{
			UserID:      userIDs[2],
			Title:       "Write unit tests",
			Description: "Increase test coverage to 80%",
			Status:      "pending",
			Priority:    "medium",
			DueDate:     timePtr(time.Now().AddDate(0, 0, 14)),
		},

		// Completed tasks
		{
			UserID:      userIDs[0],
			Title:       "Setup CI/CD pipeline",
			Description: "Configure GitHub Actions for automated deployment",
			Status:      "completed",
			Priority:    "high",
			DueDate:     timePtr(time.Now().AddDate(0, 0, -5)),
		},
		{
			UserID:      userIDs[1],
			Title:       "Create brand guidelines",
			Description: "Document brand colors, fonts, and usage",
			Status:      "completed",
			Priority:    "medium",
			DueDate:     timePtr(time.Now().AddDate(0, 0, -10)),
		},
	}

	// Insert tasks
	for _, task := range tasks {
		if err := db.DB.Create(&task).Error; err != nil {
			log.Printf("Failed to create task: %v", err)
		} else {
			fmt.Printf("✓ Created task: %s (User: %s)\n", task.Title, task.UserID)
		}
	}

	fmt.Printf("\n📊 Seeded %d tasks\n", len(tasks))
}

func timePtr(t time.Time) *time.Time {
	return &t
}

