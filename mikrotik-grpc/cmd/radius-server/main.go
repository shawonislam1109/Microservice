package main

import (
	"context"
	"isp-management-system/internal/accounting"
	"isp-management-system/internal/auth"
	"isp-management-system/internal/cache"
	"isp-management-system/internal/config"
	"isp-management-system/internal/db"
	radius_server "isp-management-system/internal/radius"
	"log"
	"os"
	"os/signal"
	"syscall"

	"layeh.com/radius"
)

func main() {
	// --- 1. Load Configuration ---
	cfg := config.Load()
	log.Println("RADIUS Server: Configuration loaded")

	// --- 2. Initialize Database and Cache ---
	repo, err := db.NewPostgresRepository(cfg.PostgresURL)
	if err != nil {
		log.Fatalf("RADIUS Server: Failed to initialize database: %v", err)
	}
	defer repo.Close()
	log.Println("RADIUS Server: Database connection successful")

	redisCache, err := cache.NewRedisCache(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	if err != nil {
		log.Fatalf("RADIUS Server: Failed to initialize redis cache: %v", err)
	}
	defer redisCache.Close()
	log.Println("RADIUS Server: Redis connection successful")

	// --- 3. Setup Accounting Worker Pool ---
	jobQueue := make(chan accounting.Job, 1000) // Buffered channel
	for i := 1; i <= cfg.WorkerPoolSize; i++ {
		go worker(i, jobQueue, repo)
	}
	log.Printf("RADIUS Server: Started %d accounting workers", cfg.WorkerPoolSize)

	// --- 4. Initialize Handlers ---
	authHandler := auth.NewHandler(repo, redisCache)
	acctHandler := accounting.NewHandler(jobQueue)

	authServerHandler := radius.HandlerFunc(authHandler.HandleAccessRequest)
	acctServerHandler := radius.HandlerFunc(acctHandler.HandleAccountingRequest)

	// --- 5. Create and Start RADIUS Server ---
	server := radius_server.NewServer(authServerHandler, acctServerHandler, cfg.RadiusSecret, cfg.AuthPort, cfg.AcctPort)

	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Fatalf("RADIUS server failed: %v", err)
		}
	}()

	log.Printf("RADIUS Server is ready. Local IP for configuration: %s", radius_server.GetLocalIP())

	// --- 6. Graceful Shutdown ---
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("RADIUS Server: Shutting down...")

	close(jobQueue)
	log.Println("RADIUS Server: Gracefully stopped")
}

// worker function for processing accounting jobs
func worker(id int, jobs <-chan accounting.Job, repo db.Repository) {
	for job := range jobs {
		log.Printf("RADIUS Worker %d: processing accounting job for user '%s'", id, radius.ToString(job.Packet.Get(radius.AttributeType("User-Name"))))
		accounting.ProcessJob(job, repo)
	}
}