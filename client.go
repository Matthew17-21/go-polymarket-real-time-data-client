package polymarketdataclient

import (
	"errors"
	"fmt"
	"net/url"
	"sync"
	"time"

	jsonutils "github.com/Matthew17-21/go-polymarket-real-time-data-client/internal/json_utils"
	"github.com/gorilla/websocket"
)

type Client interface {
	// Connect establishes a WebSocket connection to the server
	Connect() error

	// Disconnect closes the WebSocket connection
	Disconnect() error

	// Subscribe sends a subscription message to the server
	Subscribe(subscriptions []Subscription) error

	// Unsubscribe sends an unsubscription message to the server
	Unsubscribe(subscriptions []Subscription) error
}

type client struct {
	// User-provided options
	logger            Logger
	pingInterval      time.Duration
	host              string
	onConnectCallback func()
	onNewMessage      func([]byte)

	// Internal struct to store the internal state of the client
	internal struct {
		// Underlying websocket connection
		conn       *websocket.Conn
		mu         sync.RWMutex
		done       chan struct{}
		connClosed bool
	}
}

func New(opts ...ClientOptions) Client {
	c := &client{
		internal: struct {
			conn       *websocket.Conn
			mu         sync.RWMutex
			done       chan struct{}
			connClosed bool
		}{
			done: make(chan struct{}),
		},
	}

	// Set the default options
	defaultOpts()(c)

	// Apply the user-provided options
	for _, opt := range opts {
		opt(c)
	}

	// Return the client
	return c
}

// Connect establishes a WebSocket connection to the server
func (c *client) Connect() error {
	c.internal.mu.Lock()
	defer c.internal.mu.Unlock()

	// Check if the connection is closed
	if c.internal.connClosed {
		return fmt.Errorf("client is closed")
	}

	// Parse the host URL
	c.logger.Debug("Attempting to parse the host URL. host_url=%q", c.host)
	u, err := url.Parse(c.host)
	if err != nil {
		return fmt.Errorf("invalid host URL: %w", err)
	}

	// Dial the WebSocket connection
	c.logger.Debug("Attempting to dial the WebSocket connection. host_obj=%+v host_url_string=%q", u, u.String())
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	c.internal.conn = conn

	// Call the onConnectCallback if it is set
	if c.onConnectCallback != nil {
		c.onConnectCallback()
	}

	// Start background tasks
	go c.readMessages()
	go c.pingScheduler()
	c.logger.Debug("Connection established successfully!")
	return nil
}

// Disconnect closes the WebSocket connection
func (c *client) Disconnect() error {
	c.internal.mu.Lock()
	defer c.internal.mu.Unlock()

	// Check if already closed
	if c.internal.connClosed {
		return errors.New("connection is nil")
	}

	// Check if the connection is nil
	if c.internal.conn == nil {
		return errors.New("connection is nil")
	}

	// Update the flags first
	c.internal.connClosed = true
	close(c.internal.done)

	c.logger.Debug("Attempting to close the connection...")
	err := c.internal.conn.Close()
	c.internal.conn = nil // Set to nil after closing

	if err != nil {
		c.logger.Error("error closing the connection: %v", err)
		return fmt.Errorf("error while closing the connection: %w", err)
	}
	c.logger.Debug("Connection closed successfully!")
	return nil
}

// subscriptionMessage is the message sent to the server to subscribe to a topic
type subscriptionMessage struct {
	Action        string         `json:"action"`
	Subscriptions []Subscription `json:"subscriptions"`

	// TODO: Add optional clob_auth and gamma_auth
}
type Subscription struct {
	Topic   Topic       `json:"topic"`
	Type    MessageType `json:"type"`
	Filters string      `json:"filters,omitempty"`
}

// Subscribe sends a subscription message to the server
func (c *client) Subscribe(subscriptions []Subscription) error {
	c.internal.mu.RLock()
	defer c.internal.mu.RUnlock()

	if c.internal.conn == nil {
		return fmt.Errorf("not connected")
	}

	// Create the subscription message
	subscribeMsg := subscriptionMessage{
		Action:        "subscribe",
		Subscriptions: subscriptions,
	}

	// Send the subscription message
	c.logger.Debug("Attempting to send subscription message. payload=%+v", subscribeMsg)
	err := c.internal.conn.WriteJSON(subscribeMsg)
	if err != nil {
		c.logger.Error("error sending subscription message: %v", err)
		return fmt.Errorf("error while sending subscription message: %w", err)
	}
	c.logger.Debug("Subscription message sent successfully!")
	return nil
}

// Unsubscribe sends an unsubscription message to the server
func (c *client) Unsubscribe(subscriptions []Subscription) error {
	c.internal.mu.RLock()
	defer c.internal.mu.RUnlock()

	// Check if the connection is nil
	if c.internal.conn == nil {
		return fmt.Errorf("not connected")
	}

	// Create the unsubscribe message
	unsubscribeMsg := subscriptionMessage{
		Action:        "unsubscribe",
		Subscriptions: subscriptions,
	}

	// Send the unsubscribe message
	c.logger.Debug("Attempting to send unsubscribe message. payload=%+v", unsubscribeMsg)
	err := c.internal.conn.WriteJSON(unsubscribeMsg)
	if err != nil {
		c.logger.Error("error sending unsubscribe message: %v", err)
		return fmt.Errorf("error while sending unsubscribe message: %w", err)
	}
	c.logger.Debug("Unsubscribe message sent successfully!")
	return nil
}

// pingScheduler sends periodic ping messages to keep the connection alive
func (c *client) pingScheduler() {
	ticker := time.NewTicker(c.pingInterval)
	defer ticker.Stop()
	for {
		select {
		// If the client is closed, stop the ping scheduler
		case <-c.internal.done:
			return

		// If the ticker is triggered, send a ping message
		case <-ticker.C:
			if err := c.sendPing(); err != nil {
				c.logger.Error("error sending ping message: %v", err)
				return // Stop the ping scheduler
			}
		}
	}
}

// sendPing sends a `ping` message to the server
func (c *client) sendPing() error {
	// Lock the mutex to prevent race conditions
	c.internal.mu.RLock()
	defer c.internal.mu.RUnlock()

	// If the connection is nil, skip the ping
	if c.internal.conn == nil {
		return errors.New("connection is nil")
	}

	// If the connection is not nil, send a ping message
	c.logger.Debug("Attempting to send ping message...")
	err := c.internal.conn.WriteMessage(websocket.TextMessage, []byte("ping"))
	if err != nil {
		c.logger.Error("error sending ping message: %v", err)
		return fmt.Errorf("error while sending ping message: %w", err)
	}
	c.logger.Debug("Ping message sent successfully!")
	return nil
}

// readMessages handles incoming WebSocket messages
func (c *client) readMessages() {
	defer func() {
		// No cleanup needed here since Disconnect() handles closing
		c.logger.Debug("readMessages goroutine exiting")
	}()

	for {
		select {
		case <-c.internal.done:
			return
		default: // Have a default case to prevent blocking
			// Read the message
			c.internal.mu.RLock()
			conn := c.internal.conn
			c.internal.mu.RUnlock()

			if conn == nil {
				c.logger.Debug("Connection is nil. Exiting......")
				return
			}

			messageType, messageBytes, err := conn.ReadMessage()
			if err != nil {
				c.logger.Error("read error: %v", err)
				return
			}
			messageStr := string(messageBytes)
			c.logger.Debug("Received message. message_type=%d message_str=%q", messageType, messageStr)

			// Check if the message is a pong, if so, skip it
			if messageType == websocket.PongMessage || (messageType == websocket.TextMessage && messageStr == "ping") {
				c.logger.Debug("Received pong message. Skipping...")
				continue
			}

			// Check if the message is a valid JSON format.
			// In Polymarket's example, they check if the string "payload" is in the message.
			// However, we will check if the message is a valid JSON format.
			if !jsonutils.IsJsonFormat(messageStr) {
				c.logger.Debug("Received invalid JSON message. Skipping...")
				continue
			}
			c.logger.Debug("Received valid JSON message.")

			// Check if the onNewMessage callback is set.
			if c.onNewMessage == nil {
				c.logger.Debug("onNewMessage callback is not set. Skipping...")
				continue
			}
			c.logger.Debug("Sending to onNewMessage callback...")
			c.onNewMessage(messageBytes)
			c.logger.Debug("onNewMessage callback sent successfully!")
		}
	}
}
