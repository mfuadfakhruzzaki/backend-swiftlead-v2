package websocket

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// MessageType represents the type of WebSocket message
type MessageType string

const (
	MessageTypeSensorData  MessageType = "sensor_data"
	MessageTypeAlert       MessageType = "alert"
	MessageTypeNodeStatus  MessageType = "node_status"
	MessageTypePing        MessageType = "ping"
	MessageTypePong        MessageType = "pong"
	MessageTypeSubscribe   MessageType = "subscribe"
	MessageTypeUnsubscribe MessageType = "unsubscribe"
)

// Message represents a WebSocket message
type Message struct {
	Type      MessageType     `json:"type"`
	Topic     string          `json:"topic,omitempty"`
	Data      json.RawMessage `json:"data,omitempty"`
	Timestamp time.Time       `json:"timestamp"`
}

// Client represents a WebSocket client
type Client struct {
	ID            string
	UserID        string
	Conn          *websocket.Conn
	Hub           *Hub
	Send          chan []byte
	Subscriptions map[string]bool
	mu            sync.RWMutex
}

// Hub manages all WebSocket connections
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	topics     map[string]map[*Client]bool
	mu         sync.RWMutex
}

// NewHub creates a new Hub
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		topics:     make(map[string]map[*Client]bool),
	}
}

// Run starts the hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Printf("[WS] Client connected: %s (user: %s)", client.ID, client.UserID)

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
				// Remove from all topics
				for topic := range client.Subscriptions {
					if clients, ok := h.topics[topic]; ok {
						delete(clients, client)
					}
				}
			}
			h.mu.Unlock()
			log.Printf("[WS] Client disconnected: %s", client.ID)

		case message := <-h.broadcast:
			// Optimization: Do not hold the lock while sending.
			// Instead, snapshot the clients list (or just the Send channels)
			// and send in a non-blocking way or separate goroutines.
			// For simplicity and safety in Go, we can iterate with RLock but use a non-blocking send.
			// However, the original code used a non-blocking send inside the loop.
			// The issue identified was holding the lock during the iteration of potentially 10k clients.
			//
			// Improved strategy: Copy the list of clients to slice, release lock, then iterate.
			h.mu.RLock()
			clients := make([]*Client, 0, len(h.clients))
			for client := range h.clients {
				clients = append(clients, client)
			}
			h.mu.RUnlock()

			for _, client := range clients {
				select {
				case client.Send <- message:
				default:
					// Buffer full, we can drop the message or close the connection.
					// Dropping is safer for the system than blocking.
					// Closing the connection is aggressive but prevents slow consumers from leaking.
					// We will log and disconnect the slow client.
					log.Printf("[WS] Client %s buffer full, disconnecting", client.ID)
					// Handle unregistration in a non-blocking way
					go func(c *Client) {
						h.unregister <- c
					}(client)
				}
			}
		}
	}
}

// Subscribe adds a client to a topic
func (h *Hub) Subscribe(client *Client, topic string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.topics[topic]; !ok {
		h.topics[topic] = make(map[*Client]bool)
	}
	h.topics[topic][client] = true

	client.mu.Lock()
	client.Subscriptions[topic] = true
	client.mu.Unlock()

	log.Printf("[WS] Client %s subscribed to topic: %s", client.ID, topic)
}

// Unsubscribe removes a client from a topic
func (h *Hub) Unsubscribe(client *Client, topic string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if clients, ok := h.topics[topic]; ok {
		delete(clients, client)
	}

	client.mu.Lock()
	delete(client.Subscriptions, topic)
	client.mu.Unlock()

	log.Printf("[WS] Client %s unsubscribed from topic: %s", client.ID, topic)
}

// PublishToTopic sends a message to all clients subscribed to a topic
func (h *Hub) PublishToTopic(topic string, msg Message) {
	// Optimization: Same strategy as broadcast
	h.mu.RLock()
	clientsMap, ok := h.topics[topic]
	if !ok {
		h.mu.RUnlock()
		return
	}
	// Copy clients
	clients := make([]*Client, 0, len(clientsMap))
	for client := range clientsMap {
		clients = append(clients, client)
	}
	h.mu.RUnlock()

	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("[WS] Error marshaling message: %v", err)
		return
	}

	for _, client := range clients {
		select {
		case client.Send <- data:
		default:
			// Skip slow client
		}
	}
}

// BroadcastAll sends a message to all connected clients
func (h *Hub) BroadcastAll(msg Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("[WS] Error marshaling message: %v", err)
		return
	}
	h.broadcast <- data
}

// GetClientCount returns the number of connected clients
func (h *Hub) GetClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// GetTopicSubscriberCount returns the number of subscribers for a topic
func (h *Hub) GetTopicSubscriberCount(topic string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if clients, ok := h.topics[topic]; ok {
		return len(clients)
	}
	return 0
}
