package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nicolas/dirtcloud/api"
	"github.com/nicolas/dirtcloud/service"
	"github.com/nicolas/dirtcloud/service/chaos"
	"github.com/nicolas/dirtcloud/storage/sqlite"
)

func main() {
	// Load configuration from environment variables
	config := loadConfig()

	// Initialize database
	db, err := sqlite.NewDB(config.SQLiteDSN)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Initialize repositories
	projectRepo := sqlite.NewProjectRepository(db)
	instanceRepo := sqlite.NewInstanceRepository(db)
	metadataRepo := sqlite.NewMetadataRepository(db)

	// Initialize service layer
	svc := service.NewService(projectRepo, instanceRepo, metadataRepo)

	// Initialize chaos service
	chaosService := chaos.NewChaosService()

	// Initialize API handlers
	handler := api.NewHandler(svc, chaosService, config.Token)

	// Setup router
	router := api.SetupRouter(handler)

	// Create HTTP server
	server := &http.Server{
		Addr:         config.HTTPAddr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	serverErrors := make(chan error, 1)
	go func() {
		log.Printf("DirtCloud server starting on %s", config.HTTPAddr)
		serverErrors <- server.ListenAndServe()
	}()

	// Wait for shutdown signal
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		log.Fatalf("Server error: %v", err)
	case sig := <-shutdown:
		log.Printf("Received signal %v, starting graceful shutdown", sig)

		// Give outstanding requests 30 seconds to complete
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Printf("Graceful shutdown failed: %v", err)
			if err := server.Close(); err != nil {
				log.Printf("Force close failed: %v", err)
			}
		}
	}

	log.Println("Server stopped")
}

// Config holds server configuration
type Config struct {
	HTTPAddr  string
	Token     string
	SQLiteDSN string
}

// loadConfig loads configuration from environment variables
func loadConfig() Config {
	return Config{
		HTTPAddr:  getEnv("DIRT_HTTP_ADDR", ":8080"),
		Token:     getEnv("DIRT_TOKEN", ""),
		SQLiteDSN: getEnv("DIRT_SQLITE_DSN", ""),
	}
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}