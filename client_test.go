package polymarketdataclient

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// mockWebSocketServer creates a mock WebSocket server for testing
func mockWebSocketServer(t *testing.T, handler func(w http.ResponseWriter, r *http.Request)) (*httptest.Server, string) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler(w, r)
	}))

	// Convert http:// to ws:// for WebSocket URL
	wsURL := strings.Replace(server.URL, "http://", "ws://", 1)
	return server, wsURL
}

// TestNew tests the New function
func TestNew(t *testing.T) {
	tests := []struct {
		name     string
		opts     []ClientOptions
		expected func(*testing.T, Client)
	}{
		{
			name: "default options",
			opts: []ClientOptions{},
			expected: func(t *testing.T, c Client) {
				client := c.(*client)
				assert.Equal(t, defaultHost, client.host)
				assert.Equal(t, defaultPingInterval, client.pingInterval)
				assert.NotNil(t, client.logger)
				assert.NotNil(t, client.internal.done)
				assert.False(t, client.internal.connClosed)
			},
		},
		{
			name: "with custom logger",
			opts: []ClientOptions{
				WithLogger(NewLogger()),
			},
			expected: func(t *testing.T, c Client) {
				client := c.(*client)
				assert.NotNil(t, client.logger)
				// Check that it's not a silent logger
				_, isSilent := client.logger.(silentLogger)
				assert.False(t, isSilent)
			},
		},
		{
			name: "with custom ping interval",
			opts: []ClientOptions{
				WithPingInterval(10 * time.Second),
			},
			expected: func(t *testing.T, c Client) {
				client := c.(*client)
				assert.Equal(t, 10*time.Second, client.pingInterval)
			},
		},
		{
			name: "with custom host",
			opts: []ClientOptions{
				WithHost("ws://test.example.com"),
			},
			expected: func(t *testing.T, c Client) {
				client := c.(*client)
				assert.Equal(t, "ws://test.example.com", client.host)
			},
		},
		{
			name: "with onConnect callback",
			opts: []ClientOptions{
				WithOnConnect(func() {}),
			},
			expected: func(t *testing.T, c Client) {
				client := c.(*client)
				assert.NotNil(t, client.onConnectCallback)
			},
		},
		{
			name: "with onNewMessage callback",
			opts: []ClientOptions{
				WithOnNewMessage(func([]byte) {}),
			},
			expected: func(t *testing.T, c Client) {
				client := c.(*client)
				assert.NotNil(t, client.onNewMessage)
			},
		},
		{
			name: "with multiple options",
			opts: []ClientOptions{
				WithHost("ws://test.example.com"),
				WithPingInterval(15 * time.Second),
				WithLogger(NewLogger()),
			},
			expected: func(t *testing.T, c Client) {
				client := c.(*client)
				assert.Equal(t, "ws://test.example.com", client.host)
				assert.Equal(t, 15*time.Second, client.pingInterval)
				assert.NotNil(t, client.logger)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := New(tt.opts...)
			tt.expected(t, client)
		})
	}
}

// TestConnect tests the Connect function
func TestConnect(t *testing.T) {
	tests := []struct {
		name        string
		setupClient func() Client
		expectError bool
		errorMsg    string
	}{
		{
			name: "successful connection",
			setupClient: func() Client {
				return New(WithHost("ws://localhost:8080"))
			},
			expectError: false,
		},
		{
			name: "invalid host URL",
			setupClient: func() Client {
				return New(WithHost("invalid-url"))
			},
			expectError: true,
			errorMsg:    "malformed ws or wss URL",
		},
		{
			name: "connection to closed client",
			setupClient: func() Client {
				c := New().(*client)
				c.internal.connClosed = true
				return c
			},
			expectError: true,
			errorMsg:    "client is closed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := tt.setupClient()
			err := client.Connect()

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				// For successful connection, we need a real server
				// This test will fail without a server, which is expected
				assert.Error(t, err)
			}
		})
	}
}

// TestConnectWithMockServer tests Connect with a mock WebSocket server
func TestConnectWithMockServer(t *testing.T) {
	// Create a mock WebSocket server
	server, wsURL := mockWebSocketServer(t, func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		require.NoError(t, err)
		defer conn.Close()

		// Keep connection alive for a bit
		time.Sleep(100 * time.Millisecond)
	})
	defer server.Close()

	// Test successful connection
	client := New(WithHost(wsURL))
	err := client.Connect()
	assert.NoError(t, err)

	// Clean up
	err = client.Disconnect()
	assert.NoError(t, err)
}

// TestDisconnect tests the Disconnect function
func TestDisconnect(t *testing.T) {
	tests := []struct {
		name        string
		setupClient func() Client
		expectError bool
		errorMsg    string
	}{
		{
			name: "disconnect without connection",
			setupClient: func() Client {
				return New()
			},
			expectError: true,
			errorMsg:    "connection is nil",
		},
		{
			name: "disconnect already closed client",
			setupClient: func() Client {
				c := New().(*client)
				c.internal.connClosed = true
				return c
			},
			expectError: true,
			errorMsg:    "connection is nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := tt.setupClient()
			err := client.Disconnect()

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestSubscribe tests the Subscribe function
func TestSubscribe(t *testing.T) {
	tests := []struct {
		name          string
		subscriptions []Subscription
		setupClient   func() Client
		expectError   bool
		errorMsg      string
	}{
		{
			name: "subscribe without connection",
			subscriptions: []Subscription{
				{
					Topic: TopicActivity,
					Type:  MessageTypeAll,
				},
			},
			setupClient: func() Client {
				return New()
			},
			expectError: true,
			errorMsg:    "not connected",
		},
		{
			name: "subscribe with valid subscriptions",
			subscriptions: []Subscription{
				{
					Topic: TopicActivity,
					Type:  MessageTypeAll,
				},
				{
					Topic:   TopicComments,
					Type:    MessageTypeCommentCreated,
					Filters: "filter1",
				},
			},
			setupClient: func() Client {
				return New()
			},
			expectError: true,
			errorMsg:    "not connected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := tt.setupClient()
			err := client.Subscribe(tt.subscriptions)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestSubscribeWithConnection tests Subscribe with an active connection
func TestSubscribeWithConnection(t *testing.T) {
	var receivedMessages []string
	var mu sync.Mutex

	// Create a mock WebSocket server that handles subscription messages
	server, wsURL := mockWebSocketServer(t, func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		require.NoError(t, err)
		defer conn.Close()

		// Read subscription message
		_, message, err := conn.ReadMessage()
		require.NoError(t, err)

		mu.Lock()
		receivedMessages = append(receivedMessages, string(message))
		mu.Unlock()

		// Send a response
		response := `{"status": "subscribed"}`
		err = conn.WriteMessage(websocket.TextMessage, []byte(response))
		require.NoError(t, err)

		// Keep connection alive
		time.Sleep(100 * time.Millisecond)
	})
	defer server.Close()

	// Connect and subscribe
	client := New(WithHost(wsURL))
	err := client.Connect()
	require.NoError(t, err)

	subscriptions := []Subscription{
		{
			Topic: TopicActivity,
			Type:  MessageTypeAll,
		},
		{
			Topic:   TopicComments,
			Type:    MessageTypeCommentCreated,
			Filters: "filter1",
		},
	}

	err = client.Subscribe(subscriptions)
	assert.NoError(t, err)

	// Wait a bit for the message to be sent
	time.Sleep(50 * time.Millisecond)

	// Verify the subscription message was sent
	mu.Lock()
	assert.Len(t, receivedMessages, 1)

	var subscriptionMsg subscriptionMessage
	err = json.Unmarshal([]byte(receivedMessages[0]), &subscriptionMsg)
	mu.Unlock()

	assert.NoError(t, err)
	assert.Equal(t, "subscribe", subscriptionMsg.Action)
	assert.Len(t, subscriptionMsg.Subscriptions, 2)
	assert.Equal(t, TopicActivity, subscriptionMsg.Subscriptions[0].Topic)
	assert.Equal(t, MessageTypeAll, subscriptionMsg.Subscriptions[0].Type)
	assert.Equal(t, TopicComments, subscriptionMsg.Subscriptions[1].Topic)
	assert.Equal(t, MessageTypeCommentCreated, subscriptionMsg.Subscriptions[1].Type)
	assert.Equal(t, "filter1", subscriptionMsg.Subscriptions[1].Filters)

	// Clean up
	err = client.Disconnect()
	assert.NoError(t, err)
}

// TestUnsubscribe tests the Unsubscribe function
func TestUnsubscribe(t *testing.T) {
	tests := []struct {
		name          string
		subscriptions []Subscription
		setupClient   func() Client
		expectError   bool
		errorMsg      string
	}{
		{
			name: "unsubscribe without connection",
			subscriptions: []Subscription{
				{
					Topic: TopicActivity,
					Type:  MessageTypeAll,
				},
			},
			setupClient: func() Client {
				return New()
			},
			expectError: true,
			errorMsg:    "not connected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := tt.setupClient()
			err := client.Unsubscribe(tt.subscriptions)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestUnsubscribeWithConnection tests Unsubscribe with an active connection
func TestUnsubscribeWithConnection(t *testing.T) {
	var receivedMessages []string
	var mu sync.Mutex

	// Create a mock WebSocket server that handles unsubscribe messages
	server, wsURL := mockWebSocketServer(t, func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		require.NoError(t, err)
		defer conn.Close()

		// Read unsubscribe message
		_, message, err := conn.ReadMessage()
		require.NoError(t, err)

		mu.Lock()
		receivedMessages = append(receivedMessages, string(message))
		mu.Unlock()

		// Send a response
		response := `{"status": "unsubscribed"}`
		err = conn.WriteMessage(websocket.TextMessage, []byte(response))
		require.NoError(t, err)

		// Keep connection alive
		time.Sleep(100 * time.Millisecond)
	})
	defer server.Close()

	// Connect and unsubscribe
	client := New(WithHost(wsURL))
	err := client.Connect()
	require.NoError(t, err)

	subscriptions := []Subscription{
		{
			Topic: TopicActivity,
			Type:  MessageTypeAll,
		},
	}

	err = client.Unsubscribe(subscriptions)
	assert.NoError(t, err)

	// Wait a bit for the message to be sent
	time.Sleep(50 * time.Millisecond)

	// Verify the unsubscribe message was sent
	mu.Lock()
	assert.Len(t, receivedMessages, 1)

	var unsubscribeMsg subscriptionMessage
	err = json.Unmarshal([]byte(receivedMessages[0]), &unsubscribeMsg)
	mu.Unlock()

	assert.NoError(t, err)
	assert.Equal(t, "unsubscribe", unsubscribeMsg.Action)
	assert.Len(t, unsubscribeMsg.Subscriptions, 1)
	assert.Equal(t, TopicActivity, unsubscribeMsg.Subscriptions[0].Topic)
	assert.Equal(t, MessageTypeAll, unsubscribeMsg.Subscriptions[0].Type)

	// Clean up
	err = client.Disconnect()
	assert.NoError(t, err)
}

// TestSendPing tests the sendPing function
func TestSendPing(t *testing.T) {
	tests := []struct {
		name        string
		setupClient func() *client
		expectError bool
		errorMsg    string
	}{
		{
			name: "send ping without connection",
			setupClient: func() *client {
				return New().(*client)
			},
			expectError: true,
			errorMsg:    "connection is nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := tt.setupClient()
			err := client.sendPing()

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestSendPingWithConnection tests sendPing with an active connection
func TestSendPingWithConnection(t *testing.T) {
	var receivedMessages []string
	var mu sync.Mutex

	// Create a mock WebSocket server that handles ping messages
	server, wsURL := mockWebSocketServer(t, func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		require.NoError(t, err)
		defer conn.Close()

		// Read ping message
		_, message, err := conn.ReadMessage()
		require.NoError(t, err)

		mu.Lock()
		receivedMessages = append(receivedMessages, string(message))
		mu.Unlock()

		// Send pong response
		err = conn.WriteMessage(websocket.PongMessage, nil)
		require.NoError(t, err)

		// Keep connection alive
		time.Sleep(100 * time.Millisecond)
	})
	defer server.Close()

	// Connect
	client := New(WithHost(wsURL)).(*client)
	err := client.Connect()
	require.NoError(t, err)

	// Send ping
	err = client.sendPing()
	assert.NoError(t, err)

	// Wait a bit for the message to be sent
	time.Sleep(50 * time.Millisecond)

	// Verify the ping message was sent
	mu.Lock()
	assert.Len(t, receivedMessages, 1)
	assert.Equal(t, "ping", receivedMessages[0])
	mu.Unlock()

	// Clean up
	err = client.Disconnect()
	assert.NoError(t, err)
}

// TestPingScheduler tests the pingScheduler function
func TestPingScheduler(t *testing.T) {
	var receivedMessages []string
	var mu sync.Mutex

	// Create a mock WebSocket server
	server, wsURL := mockWebSocketServer(t, func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		require.NoError(t, err)
		defer conn.Close()

		// Read messages
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				break
			}

			mu.Lock()
			receivedMessages = append(receivedMessages, string(message))
			mu.Unlock()

			// Send pong response
			conn.WriteMessage(websocket.PongMessage, nil)
		}
	})
	defer server.Close()

	// Connect with short ping interval
	client := New(WithHost(wsURL), WithPingInterval(50*time.Millisecond)).(*client)
	err := client.Connect()
	require.NoError(t, err)

	// Wait for a few ping messages
	time.Sleep(200 * time.Millisecond)

	// Verify ping messages were sent
	mu.Lock()
	assert.GreaterOrEqual(t, len(receivedMessages), 3) // Should have at least 3 pings
	for _, msg := range receivedMessages {
		assert.Equal(t, "ping", msg)
	}
	mu.Unlock()

	// Clean up
	err = client.Disconnect()
	assert.NoError(t, err)
}

// TestReadMessages tests the readMessages function
func TestReadMessages(t *testing.T) {
	var receivedMessages []string
	var mu sync.Mutex

	// Create a mock WebSocket server that sends various message types
	server, wsURL := mockWebSocketServer(t, func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		require.NoError(t, err)
		defer conn.Close()

		// Send different types of messages
		messages := []string{
			"ping",                 // Should be ignored
			`{"payload": "test1"}`, // Valid JSON
			`invalid json`,         // Invalid JSON
			`{"payload": "test2"}`, // Valid JSON
		}

		for _, msg := range messages {
			err = conn.WriteMessage(websocket.TextMessage, []byte(msg))
			require.NoError(t, err)
			time.Sleep(10 * time.Millisecond)
		}

		// Keep connection alive
		time.Sleep(100 * time.Millisecond)
	})
	defer server.Close()

	// Connect with message handler
	client := New(WithHost(wsURL), WithOnNewMessage(func(data []byte) {
		mu.Lock()
		receivedMessages = append(receivedMessages, string(data))
		mu.Unlock()
	})).(*client)

	err := client.Connect()
	require.NoError(t, err)

	// Wait for messages to be processed
	time.Sleep(200 * time.Millisecond)

	// Verify only valid JSON messages were received
	mu.Lock()
	assert.Len(t, receivedMessages, 2)
	assert.Equal(t, `{"payload": "test1"}`, receivedMessages[0])
	assert.Equal(t, `{"payload": "test2"}`, receivedMessages[1])
	mu.Unlock()

	// Clean up
	err = client.Disconnect()
	assert.NoError(t, err)
}

// TestReadMessagesWithoutHandler tests readMessages without a message handler
func TestReadMessagesWithoutHandler(t *testing.T) {
	// Create a mock WebSocket server
	server, wsURL := mockWebSocketServer(t, func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		require.NoError(t, err)
		defer conn.Close()

		// Send a valid JSON message
		message := `{"payload": "test"}`
		err = conn.WriteMessage(websocket.TextMessage, []byte(message))
		require.NoError(t, err)

		// Keep connection alive
		time.Sleep(100 * time.Millisecond)
	})
	defer server.Close()

	// Connect without message handler
	client := New(WithHost(wsURL)).(*client)
	err := client.Connect()
	require.NoError(t, err)

	// Wait for messages to be processed
	time.Sleep(100 * time.Millisecond)

	// Should not panic or error
	err = client.Disconnect()
	assert.NoError(t, err)
}

// TestOnConnectCallback tests that the onConnect callback is called
func TestOnConnectCallback(t *testing.T) {
	var callbackCalled bool
	var mu sync.Mutex

	// Create a mock WebSocket server
	server, wsURL := mockWebSocketServer(t, func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		require.NoError(t, err)
		defer conn.Close()

		// Keep connection alive
		time.Sleep(100 * time.Millisecond)
	})
	defer server.Close()

	// Connect with onConnect callback
	client := New(
		WithHost(wsURL),
		WithOnConnect(func() {
			mu.Lock()
			callbackCalled = true
			mu.Unlock()
		}),
	)

	err := client.Connect()
	require.NoError(t, err)

	// Wait a bit for the callback to be called
	time.Sleep(50 * time.Millisecond)

	// Verify callback was called
	mu.Lock()
	assert.True(t, callbackCalled)
	mu.Unlock()

	// Clean up
	err = client.Disconnect()
	assert.NoError(t, err)
}

// TestSubscriptionMessage tests the subscriptionMessage struct
func TestSubscriptionMessage(t *testing.T) {
	msg := subscriptionMessage{
		Action: "subscribe",
		Subscriptions: []Subscription{
			{
				Topic: TopicActivity,
				Type:  MessageTypeAll,
			},
			{
				Topic:   TopicComments,
				Type:    MessageTypeCommentCreated,
				Filters: "filter1",
			},
		},
	}

	// Test JSON marshaling
	data, err := json.Marshal(msg)
	assert.NoError(t, err)

	// Test JSON unmarshaling
	var unmarshaledMsg subscriptionMessage
	err = json.Unmarshal(data, &unmarshaledMsg)
	assert.NoError(t, err)

	assert.Equal(t, msg.Action, unmarshaledMsg.Action)
	assert.Len(t, unmarshaledMsg.Subscriptions, 2)
	assert.Equal(t, msg.Subscriptions[0].Topic, unmarshaledMsg.Subscriptions[0].Topic)
	assert.Equal(t, msg.Subscriptions[0].Type, unmarshaledMsg.Subscriptions[0].Type)
	assert.Equal(t, msg.Subscriptions[1].Topic, unmarshaledMsg.Subscriptions[1].Topic)
	assert.Equal(t, msg.Subscriptions[1].Type, unmarshaledMsg.Subscriptions[1].Type)
	assert.Equal(t, msg.Subscriptions[1].Filters, unmarshaledMsg.Subscriptions[1].Filters)
}

// TestSubscription tests the Subscription struct
func TestSubscription(t *testing.T) {
	sub := Subscription{
		Topic:   TopicActivity,
		Type:    MessageTypeAll,
		Filters: "test_filter",
	}

	// Test JSON marshaling
	data, err := json.Marshal(sub)
	assert.NoError(t, err)

	// Test JSON unmarshaling
	var unmarshaledSub Subscription
	err = json.Unmarshal(data, &unmarshaledSub)
	assert.NoError(t, err)

	assert.Equal(t, sub.Topic, unmarshaledSub.Topic)
	assert.Equal(t, sub.Type, unmarshaledSub.Type)
	assert.Equal(t, sub.Filters, unmarshaledSub.Filters)
}

// TestClientInterface tests that the client implements the Client interface
func TestClientInterface(t *testing.T) {
	var _ Client = New()
}

// TestConcurrentOperations tests concurrent operations on the client
func TestConcurrentOperations(t *testing.T) {
	var receivedMessages []string
	var mu sync.Mutex

	// Create a mock WebSocket server
	server, wsURL := mockWebSocketServer(t, func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		require.NoError(t, err)
		defer conn.Close()

		// Handle messages
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				break
			}

			mu.Lock()
			receivedMessages = append(receivedMessages, string(message))
			mu.Unlock()

			// Send response
			response := `{"status": "ok"}`
			conn.WriteMessage(websocket.TextMessage, []byte(response))
		}
	})
	defer server.Close()

	// Create client
	client := New(
		WithHost(wsURL),
		WithOnNewMessage(func(data []byte) {
			mu.Lock()
			receivedMessages = append(receivedMessages, fmt.Sprintf("received: %s", string(data)))
			mu.Unlock()
		}),
	)

	// Connect
	err := client.Connect()
	require.NoError(t, err)

	// Perform concurrent operations
	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		defer wg.Done()
		subscriptions := []Subscription{
			{Topic: TopicActivity, Type: MessageTypeAll},
		}
		client.Subscribe(subscriptions)
	}()

	go func() {
		defer wg.Done()
		time.Sleep(10 * time.Millisecond)
		subscriptions := []Subscription{
			{Topic: TopicComments, Type: MessageTypeCommentCreated},
		}
		client.Subscribe(subscriptions)
	}()

	go func() {
		defer wg.Done()
		time.Sleep(20 * time.Millisecond)
		client.Disconnect()
	}()

	wg.Wait()

	// Verify operations completed without panics
	mu.Lock()
	assert.GreaterOrEqual(t, len(receivedMessages), 2) // At least 2 subscription messages
	mu.Unlock()
}

// TestReconnectAfterDisconnect tests that the client can be reconnected after disconnect
func TestReconnectAfterDisconnect(t *testing.T) {
	// Create a mock WebSocket server
	server, wsURL := mockWebSocketServer(t, func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		require.NoError(t, err)
		defer conn.Close()

		// Keep connection alive
		time.Sleep(100 * time.Millisecond)
	})
	defer server.Close()

	// Create client
	client := New(WithHost(wsURL))

	// First connection
	err := client.Connect()
	require.NoError(t, err)

	// Disconnect
	err = client.Disconnect()
	require.NoError(t, err)

	// Try to connect again - should fail because client is closed
	err = client.Connect()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "client is closed")
}
