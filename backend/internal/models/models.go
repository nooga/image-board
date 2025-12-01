package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Topic represents an image post that starts a discussion thread
type Topic struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Title     string             `json:"title" bson:"title"`
	ImageURL  string             `json:"imageUrl" bson:"imageUrl"`
	ImageKey  string             `json:"-" bson:"imageKey"`
	Author    string             `json:"author" bson:"author"`
	CreatedAt time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt time.Time          `json:"updatedAt" bson:"updatedAt"`
}

// Message represents a chat message within a topic
type Message struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	TopicID   primitive.ObjectID `json:"topicId" bson:"topicId"`
	Content   string             `json:"content" bson:"content"`
	Author    string             `json:"author" bson:"author"`
	CreatedAt time.Time          `json:"createdAt" bson:"createdAt"`
}

// TopicWithMessages combines a topic with its messages for API responses
type TopicWithMessages struct {
	Topic    Topic     `json:"topic"`
	Messages []Message `json:"messages"`
}

// CreateTopicRequest represents the request to create a new topic
type CreateTopicRequest struct {
	Title  string `json:"title"`
	Author string `json:"author"`
}

// CreateMessageRequest represents the request to create a new message
type CreateMessageRequest struct {
	Content string `json:"content"`
	Author  string `json:"author"`
}

// WebSocket message types
type WSMessageType string

const (
	WSTypeNewTopic   WSMessageType = "new_topic"
	WSTypeNewMessage WSMessageType = "new_message"
	WSTypeError      WSMessageType = "error"
)

// WSMessage represents a WebSocket message
type WSMessage struct {
	Type    WSMessageType `json:"type"`
	Payload interface{}   `json:"payload"`
}

// NewTopicPayload is the payload for new topic events
type NewTopicPayload struct {
	Topic Topic `json:"topic"`
}

// NewMessagePayload is the payload for new message events
type NewMessagePayload struct {
	Message Message `json:"message"`
}
