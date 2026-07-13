package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"imageboard/internal/api"
	"imageboard/internal/config"
	"imageboard/internal/pubsub"
	"imageboard/internal/repository"
	"imageboard/internal/storage"
	"imageboard/internal/websocket"
)

func main() {
	cfg := config.Load()

	// Initialize MongoDB
	repo, err := repository.NewMongoRepository(cfg.MongoURI, cfg.MongoDB)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer repo.Close()

	// Initialize Redis pub/sub
	ps, err := pubsub.NewRedisPubSub(cfg.RedisAddr)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer ps.Close()

	// Initialize S3/MinIO storage. Connection is lazy — the client is created
	// here but no network call is made until the first upload or download.
	// If MINIO_ENDPOINT is not set the default "localhost:9000" is used and
	// image operations will return errors at runtime (not at startup).
	store, err := storage.NewS3Storage(cfg.MinioEndpoint, cfg.MinioAccessKey, cfg.MinioSecretKey, cfg.MinioBucket, cfg.MinioUseSSL)
	if err != nil {
		log.Fatalf("Failed to initialize MinIO storage client: %v", err)
	}

	// Initialize WebSocket hub
	hub := websocket.NewHub(ps)
	go hub.Run()

	// Create router
	router := api.NewRouter(repo, store, hub, ps)

	// Create server
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Server starting on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped")
}
