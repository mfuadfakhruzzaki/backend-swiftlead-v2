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

// broadcastMessage is an internal struct to handle topic-based broadcasts
type broadcastMessage struct {
	topic string
	msg   Message
	all   bool // if true, broadcast to all
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
	broadcast  chan broadcastMessage
	register   chan *Client
	unregister chan *Client
	topics     map[string]map[*Client]bool
	mu         sync.RWMutex // Still needed for read-only access like GetClientCount
}

// NewHub creates a new Hub
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan broadcastMessage, 256),
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
			// Lock for consistency with external readers, though this loop owns the map write access
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
						// Clean up empty topics
						if len(clients) == 0 {
							delete(h.topics, topic)
						}
					}
				}
			}
			h.mu.Unlock()
			log.Printf("[WS] Client disconnected: %s", client.ID)

		case bMsg := <-h.broadcast:
			// Prepare data once
			data, err := json.Marshal(bMsg.msg)
			if err != nil {
				log.Printf("[WS] Error marshaling message: %v", err)
				continue
			}

			// Define target clients
			var targets map[*Client]bool

			h.mu.Lock() // Lock to safely read/write maps if needed, mainly for topic map access
			if bMsg.all {
				targets = h.clients
			} else {
				if tClients, ok := h.topics[bMsg.topic]; ok {
					targets = tClients
				}
			}

			// Iterate and send.
			// CRITICAL SAFETY: This loop is serialized with the unregister case above.
			// Therefore, we are guaranteed that 'client.Send' is open because if it were closed,
			// the client would have been removed from 'targets' (which are subset of h.clients/h.topics)
			// in the unregister case before we got here or will be processed after we finish.
			for client := range targets {
				select {
				case client.Send <- data:
				default:
					// Buffer full. In this serialized model, we must NOT block.
					// We also cannot immediately unregister by sending to h.unregister channel
					// because that would block if the channel is full (deadlock risk) or
					// create a race if we tried to call the unregister logic directly.
					//
					// Best practice: Close the channel and delete immediately.
					// Since we are IN the Run loop, we own the state.
					log.Printf("[WS] Client %s buffer full, disconnecting", client.ID)

					delete(h.clients, client)
					close(client.Send)
					// Cleanup topics
					for topic := range client.Subscriptions {
						if tClients, ok := h.topics[topic]; ok {
							delete(tClients, client)
							if len(tClients) == 0 {
								delete(h.topics, topic)
							}
						}
					}
				}
			}
			h.mu.Unlock()
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
		if len(clients) == 0 {
			delete(h.topics, topic)
		}
	}

	client.mu.Lock()
	delete(client.Subscriptions, topic)
	client.mu.Unlock()

	log.Printf("[WS] Client %s unsubscribed from topic: %s", client.ID, topic)
}

// PublishToTopic sends a message to all clients subscribed to a topic
func (h *Hub) PublishToTopic(topic string, msg Message) {
	h.broadcast <- broadcastMessage{
		topic: topic,
		msg:   msg,
		all:   false,
	}
}

// BroadcastAll sends a message to all connected clients
func (h *Hub) BroadcastAll(msg Message) {
	h.broadcast <- broadcastMessage{
		msg: msg,
		all: true,
	}
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
