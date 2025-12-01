package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"imageboard/internal/models"
	"imageboard/internal/repository"
	"imageboard/internal/storage"
	"imageboard/internal/websocket"
)

type Handlers struct {
	repo    repository.Repository
	storage storage.Storage
	hub     *websocket.Hub
}

func NewHandlers(repo repository.Repository, store storage.Storage, hub *websocket.Hub) *Handlers {
	return &Handlers{
		repo:    repo,
		storage: store,
		hub:     hub,
	}
}

func (h *Handlers) ListTopics(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	if limit <= 0 || limit > 100 {
		limit = 50
	}

	topics, err := h.repo.ListTopics(r.Context(), limit, offset)
	if err != nil {
		http.Error(w, "Failed to fetch topics", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, topics)
}

func (h *Handlers) CreateTopic(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form (max 10MB)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	title := r.FormValue("title")
	author := r.FormValue("author")

	if title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}

	if author == "" {
		author = "Anonymous"
	}

	// Get uploaded file
	file, header, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Image is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validate content type
	contentType := header.Header.Get("Content-Type")
	if !isValidImageType(contentType) {
		http.Error(w, "Invalid image type. Allowed: jpg, png, gif, webp", http.StatusBadRequest)
		return
	}

	// Upload to S3/MinIO
	key, url, err := h.storage.Upload(r.Context(), file, header.Size, contentType)
	if err != nil {
		http.Error(w, "Failed to upload image", http.StatusInternalServerError)
		return
	}

	// Create topic in database
	topic := &models.Topic{
		Title:    title,
		Author:   author,
		ImageURL: url,
		ImageKey: key,
	}

	if err := h.repo.CreateTopic(r.Context(), topic); err != nil {
		http.Error(w, "Failed to create topic", http.StatusInternalServerError)
		return
	}

	// Broadcast new topic via WebSocket
	h.hub.BroadcastNewTopic(topic)

	writeJSON(w, http.StatusCreated, topic)
}

func (h *Handlers) GetTopic(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		http.Error(w, "Invalid topic ID", http.StatusBadRequest)
		return
	}

	topic, err := h.repo.GetTopic(r.Context(), id)
	if err != nil {
		http.Error(w, "Topic not found", http.StatusNotFound)
		return
	}

	messages, err := h.repo.GetMessagesByTopic(r.Context(), id, 100, 0)
	if err != nil {
		http.Error(w, "Failed to fetch messages", http.StatusInternalServerError)
		return
	}

	response := models.TopicWithMessages{
		Topic:    *topic,
		Messages: messages,
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *Handlers) CreateMessage(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	topicID, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		http.Error(w, "Invalid topic ID", http.StatusBadRequest)
		return
	}

	// Verify topic exists
	_, err = h.repo.GetTopic(r.Context(), topicID)
	if err != nil {
		http.Error(w, "Topic not found", http.StatusNotFound)
		return
	}

	var req models.CreateMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Content == "" {
		http.Error(w, "Content is required", http.StatusBadRequest)
		return
	}

	if req.Author == "" {
		req.Author = "Anonymous"
	}

	message := &models.Message{
		TopicID: topicID,
		Content: req.Content,
		Author:  req.Author,
	}

	if err := h.repo.CreateMessage(r.Context(), message); err != nil {
		http.Error(w, "Failed to create message", http.StatusInternalServerError)
		return
	}

	// Update topic timestamp
	h.repo.UpdateTopicTimestamp(r.Context(), topicID)

	// Broadcast new message via WebSocket
	h.hub.BroadcastNewMessage(idStr, message)

	writeJSON(w, http.StatusCreated, message)
}

func (h *Handlers) HandleFeedWS(w http.ResponseWriter, r *http.Request) {
	h.hub.HandleFeedConnection(w, r)
}

func (h *Handlers) HandleTopicWS(w http.ResponseWriter, r *http.Request) {
	topicID := chi.URLParam(r, "id")
	if topicID == "" {
		http.Error(w, "Topic ID required", http.StatusBadRequest)
		return
	}
	h.hub.HandleTopicConnection(w, r, topicID)
}

func isValidImageType(contentType string) bool {
	validTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
	}
	return validTypes[contentType]
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
