package websocket

import (
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period (must be less than pongWait)
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 512
)

// NewClient creates a new WebSocket client
func NewClient(hub *Hub, conn *websocket.Conn, userID string) *Client {
	return &Client{
		ID:            uuid.New().String(),
		UserID:        userID,
		Hub:           hub,
		Conn:          conn,
		Send:          make(chan []byte, 256),
		Subscriptions: make(map[string]bool),
	}
}

// ReadPump pumps messages from the WebSocket connection to the hub
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("[WS] Error: %v", err)
			}
			break
		}

		// Parse incoming message
		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("[WS] Invalid message format: %v", err)
			continue
		}

		// Handle message types
		switch msg.Type {
		case MessageTypePing:
			c.sendPong()
		case MessageTypeSubscribe:
			if msg.Topic != "" {
				c.Hub.Subscribe(c, msg.Topic)
			}
		case MessageTypeUnsubscribe:
			if msg.Topic != "" {
				c.Hub.Unsubscribe(c, msg.Topic)
			}
		}
	}
}

// WritePump pumps messages from the hub to the WebSocket connection
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub closed the channel
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current WebSocket message
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) sendPong() {
	msg := Message{
		Type:      MessageTypePong,
		Timestamp: time.Now(),
	}
	data, _ := json.Marshal(msg)
	select {
	case c.Send <- data:
	default:
	}
}
