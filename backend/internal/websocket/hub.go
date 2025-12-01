package websocket

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"imageboard/internal/models"
	"imageboard/internal/pubsub"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins in development
	},
}

// Client represents a connected WebSocket client
type Client struct {
	hub     *Hub
	conn    *websocket.Conn
	send    chan []byte
	channel string
}

// Hub manages WebSocket connections and message broadcasting
type Hub struct {
	clients    map[*Client]bool
	channels   map[string]map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan *ChannelMessage
	pubsub     pubsub.PubSub
	mu         sync.RWMutex
}

// ChannelMessage represents a message to broadcast to a specific channel
type ChannelMessage struct {
	Channel string
	Data    []byte
}

func NewHub(ps pubsub.PubSub) *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		channels:   make(map[string]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *ChannelMessage, 256),
		pubsub:     ps,
	}
}

func (h *Hub) Run() {
	// Subscribe to the global feed channel
	ctx := context.Background()
	feedChan, cleanupFeed := h.pubsub.Subscribe(ctx, pubsub.ChannelFeed)
	defer cleanupFeed()

	go func() {
		for msg := range feedChan {
			h.broadcast <- &ChannelMessage{
				Channel: pubsub.ChannelFeed,
				Data:    msg,
			}
		}
	}()

	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			if _, ok := h.channels[client.channel]; !ok {
				h.channels[client.channel] = make(map[*Client]bool)
				// Subscribe to Redis channel for this topic
				if client.channel != pubsub.ChannelFeed {
					go h.subscribeToChannel(ctx, client.channel)
				}
			}
			h.channels[client.channel][client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				delete(h.channels[client.channel], client)
				close(client.send)
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			if clients, ok := h.channels[message.Channel]; ok {
				for client := range clients {
					select {
					case client.send <- message.Data:
					default:
						close(client.send)
						delete(h.clients, client)
						delete(h.channels[message.Channel], client)
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) subscribeToChannel(ctx context.Context, channel string) {
	msgChan, cleanup := h.pubsub.Subscribe(ctx, channel)
	defer cleanup()

	for msg := range msgChan {
		h.broadcast <- &ChannelMessage{
			Channel: channel,
			Data:    msg,
		}
	}
}

// HandleFeedConnection handles WebSocket connections for the global feed
func (h *Hub) HandleFeedConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	client := &Client{
		hub:     h,
		conn:    conn,
		send:    make(chan []byte, 256),
		channel: pubsub.ChannelFeed,
	}

	h.register <- client

	go client.writePump()
	go client.readPump()
}

// HandleTopicConnection handles WebSocket connections for a specific topic
func (h *Hub) HandleTopicConnection(w http.ResponseWriter, r *http.Request, topicID string) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	channel := h.pubsub.TopicChannel(topicID)
	client := &Client{
		hub:     h,
		conn:    conn,
		send:    make(chan []byte, 256),
		channel: channel,
	}

	h.register <- client

	go client.writePump()
	go client.readPump()
}

// BroadcastNewTopic publishes a new topic to the feed
func (h *Hub) BroadcastNewTopic(topic *models.Topic) {
	msg := models.WSMessage{
		Type: models.WSTypeNewTopic,
		Payload: models.NewTopicPayload{
			Topic: *topic,
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := h.pubsub.Publish(ctx, pubsub.ChannelFeed, msg); err != nil {
		log.Printf("Failed to publish new topic: %v", err)
	}
}

// BroadcastNewMessage publishes a new message to a topic channel
func (h *Hub) BroadcastNewMessage(topicID string, message *models.Message) {
	msg := models.WSMessage{
		Type: models.WSTypeNewMessage,
		Payload: models.NewMessagePayload{
			Message: *message,
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	channel := h.pubsub.TopicChannel(topicID)
	if err := h.pubsub.Publish(ctx, channel, msg); err != nil {
		log.Printf("Failed to publish new message: %v", err)
	}
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(512)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// Parse and re-encode the message to ensure proper JSON
			var wsMsg models.WSMessage
			if err := json.Unmarshal(message, &wsMsg); err == nil {
				if err := c.conn.WriteJSON(wsMsg); err != nil {
					return
				}
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
