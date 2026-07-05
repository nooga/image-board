package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"imageboard/internal/pubsub"
	"imageboard/internal/repository"
	"imageboard/internal/storage"
	"imageboard/internal/websocket"
)

func NewRouter(repo repository.Repository, store storage.Storage, hub *websocket.Hub, ps pubsub.PubSub) http.Handler {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173", "http://localhost:3000", "*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	handlers := NewHandlers(repo, store, hub)

	// REST API routes
	r.Route("/api", func(r chi.Router) {
		r.Get("/topics", handlers.ListTopics)
		r.Post("/topics", handlers.CreateTopic)
		r.Get("/topics/{id}", handlers.GetTopic)
		r.Post("/topics/{id}/messages", handlers.CreateMessage)
		r.Get("/images/*", handlers.ServeImage)
	})

	// WebSocket routes
	r.Get("/ws/feed", handlers.HandleFeedWS)
	r.Get("/ws/topics/{id}", handlers.HandleTopicWS)

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	return r
}
