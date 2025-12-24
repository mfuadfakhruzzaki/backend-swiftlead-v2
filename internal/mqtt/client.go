package mqtt

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	pahomqtt "github.com/eclipse/paho.mqtt.golang"

	"github.com/swiftlead/backend-swiftlet/internal/config"
	"github.com/swiftlead/backend-swiftlet/internal/models"
	"github.com/swiftlead/backend-swiftlet/pkg/logger"
)

// MessageHandler is a function that processes incoming MQTT messages
type MessageHandler func(ctx context.Context, payload *models.SensorPayload) error

// Client wraps the Paho MQTT client
type Client struct {
	client  pahomqtt.Client
	cfg     *config.Config
	handler MessageHandler
}

// NewClient creates a new MQTT client
func NewClient(cfg *config.Config) *Client {
	return &Client{
		cfg: cfg,
	}
}

// SetHandler sets the message handler for incoming sensor data
func (c *Client) SetHandler(handler MessageHandler) {
	c.handler = handler
}

// Connect establishes connection to the MQTT broker
func (c *Client) Connect() error {
	opts := pahomqtt.NewClientOptions()
	opts.AddBroker(c.cfg.MQTTBroker)
	opts.SetClientID(c.cfg.MQTTClientID)
	opts.SetKeepAlive(time.Duration(c.cfg.MQTTKeepAlive) * time.Second)
	opts.SetConnectTimeout(time.Duration(c.cfg.MQTTConnectTimeout) * time.Second)
	opts.SetAutoReconnect(true)
	opts.SetMaxReconnectInterval(30 * time.Second)

	if c.cfg.MQTTUsername != "" {
		opts.SetUsername(c.cfg.MQTTUsername)
		opts.SetPassword(c.cfg.MQTTPassword)
	}

	opts.SetOnConnectHandler(c.onConnect)
	opts.SetConnectionLostHandler(c.onConnectionLost)
	opts.SetReconnectingHandler(c.onReconnecting)

	c.client = pahomqtt.NewClient(opts)

	token := c.client.Connect()
	if token.Wait() && token.Error() != nil {
		return fmt.Errorf("failed to connect to MQTT broker: %w", token.Error())
	}

	logger.Info("Connected to MQTT broker: %s", c.cfg.MQTTBroker)
	return nil
}

// Subscribe subscribes to sensor telemetry topics
func (c *Client) Subscribe() error {
	topic := c.cfg.MQTTTopicPrefix + "/+"
	token := c.client.Subscribe(topic, 1, c.messageHandler)
	if token.Wait() && token.Error() != nil {
		return fmt.Errorf("failed to subscribe to topic %s: %w", topic, token.Error())
	}
	logger.Info("Subscribed to MQTT topic: %s", topic)
	return nil
}

// Publish publishes a message to a topic
func (c *Client) Publish(topic string, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	token := c.client.Publish(topic, 1, false, data)
	if token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

// PublishAudioCommand publishes an audio control command
func (c *Client) PublishAudioCommand(action string, value int) error {
	payload := map[string]interface{}{
		"action": action,
		"value":  value,
	}
	return c.Publish("swiftlead/cmd/lmb/set", payload)
}

// PublishPumpCommand publishes a pump control command
func (c *Client) PublishPumpCommand(value int) error {
	payload := map[string]interface{}{
		"action": "sprayer_set",
		"value":  value,
	}
	return c.Publish("swiftlead/cmd/pump/set", payload)
}

// Disconnect closes the MQTT connection
func (c *Client) Disconnect() {
	if c.client != nil && c.client.IsConnected() {
		c.client.Disconnect(1000)
		logger.Info("Disconnected from MQTT broker")
	}
}

// IsConnected checks if client is connected
func (c *Client) IsConnected() bool {
	return c.client != nil && c.client.IsConnected()
}

// Callbacks
func (c *Client) onConnect(client pahomqtt.Client) {
	logger.Info("MQTT connected, subscribing to topics...")
	if err := c.Subscribe(); err != nil {
		logger.Error("Failed to subscribe: %v", err)
	}
}

func (c *Client) onConnectionLost(client pahomqtt.Client, err error) {
	logger.Warn("MQTT connection lost: %v", err)
}

func (c *Client) onReconnecting(client pahomqtt.Client, opts *pahomqtt.ClientOptions) {
	logger.Info("MQTT reconnecting...")
}

// messageHandler processes incoming MQTT messages
func (c *Client) messageHandler(client pahomqtt.Client, msg pahomqtt.Message) {
	logger.Debug("Received MQTT message on topic: %s", msg.Topic())

	var payload models.SensorPayload
	if err := json.Unmarshal(msg.Payload(), &payload); err != nil {
		logger.Error("Failed to parse MQTT message: %v", err)
		return
	}

	if c.handler != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := c.handler(ctx, &payload); err != nil {
			logger.Error("Failed to process sensor payload: %v", err)
		}
	}
}
