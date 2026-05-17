package main

import (
	"context"
	"isp-management-system/internal/api"
	"isp-management-system/internal/config"
	"isp-management-system/internal/db"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	// --- 1. Load Configuration ---
	cfg := config.Load()
	log.Println("API Server: Configuration loaded")

	// --- 2. Initialize Database ---
	repo, err := db.NewPostgresRepository(cfg.PostgresURL)
	if err != nil {
		log.Fatalf("API Server: Failed to initialize database: %v", err)
	}
	defer repo.Close()
	log.Println("API Server: Database connection successful")

	// --- 3. Setup HTTP Server (Fiber) ---
	app := fiber.New()
	app.Use(logger.New()) // Add a request logger middleware

	// --- 4. Initialize Handler and Routes ---
	apiHandler := api.NewHandler(repo)
	apiGroup := app.Group("/api")
	apiGroup.Post("/clients", apiHandler.CreateUser)

	// --- 5. Start Server ---
	go func() {
		log.Printf("API Server: Starting on port %s", cfg.APIPort)
		if err := app.Listen(cfg.APIPort); err != nil {
			log.Fatalf("API Server: failed to start: %v", err)
		}
	}()

	// --- 6. Graceful Shutdown ---
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("API Server: Shutting down...")

	// Give the server 5 seconds to finish outstanding requests
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		log.Fatalf("API Server: Shutdown failed: %v", err)
	}

	log.Println("API Server: Gracefully stopped")
}
