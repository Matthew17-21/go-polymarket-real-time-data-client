package polymarketdataclient

import (
	"testing"
	"time"
)

func TestDefaultOpts(t *testing.T) {
	c := &client{}

	// Apply default options
	defaultOpts()(c)

	// Verify default values
	if c.host != defaultHost {
		t.Errorf("expected host to be %s, got %s", defaultHost, c.host)
	}

	if c.pingInterval != defaultPingInterval {
		t.Errorf("expected ping interval to be %v, got %v", defaultPingInterval, c.pingInterval)
	}

	if c.logger == nil {
		t.Error("expected logger to be set, got nil")
	}

}

func TestWithLogger(t *testing.T) {
	c := &client{}
	customLogger := NewLogger()

	// Apply WithLogger option
	WithLogger(customLogger)(c)

	// Verify logger is set
	if c.logger != customLogger {
		t.Error("expected logger to be set to custom logger")
	}
}

func TestWithLoggerNil(t *testing.T) {
	c := &client{}

	// Apply WithLogger option with nil
	WithLogger(nil)(c)

	// Verify logger is set to nil
	if c.logger != nil {
		t.Error("expected logger to be nil")
	}
}

func TestWithPingInterval(t *testing.T) {
	c := &client{}
	customInterval := 10 * time.Second

	// Apply WithPingInterval option
	WithPingInterval(customInterval)(c)

	// Verify ping interval is set
	if c.pingInterval != customInterval {
		t.Errorf("expected ping interval to be %v, got %v", customInterval, c.pingInterval)
	}
}

func TestWithPingIntervalZero(t *testing.T) {
	c := &client{}

	// Apply WithPingInterval option with zero duration
	WithPingInterval(0)(c)

	// Verify ping interval is set to zero
	if c.pingInterval != 0 {
		t.Errorf("expected ping interval to be 0, got %v", c.pingInterval)
	}
}

func TestWithHost(t *testing.T) {
	c := &client{}
	customHost := "wss://custom.example.com"

	// Apply WithHost option
	WithHost(customHost)(c)

	// Verify host is set
	if c.host != customHost {
		t.Errorf("expected host to be %s, got %s", customHost, c.host)
	}
}

func TestWithHostEmpty(t *testing.T) {
	c := &client{}

	// Apply WithHost option with empty string
	WithHost("")(c)

	// Verify host is set to empty string
	if c.host != "" {
		t.Errorf("expected host to be empty string, got %s", c.host)
	}
}

func TestWithOnConnect(t *testing.T) {
	c := &client{}
	called := false
	callback := func() {
		called = true
	}

	// Apply WithOnConnect option
	WithOnConnect(callback)(c)

	// Verify callback is set
	if c.onConnectCallback == nil {
		t.Error("expected onConnectCallback to be set")
	}

	// Test that the callback works
	c.onConnectCallback()
	if !called {
		t.Error("expected callback to be called")
	}
}

func TestWithOnConnectNil(t *testing.T) {
	c := &client{}

	// Apply WithOnConnect option with nil
	WithOnConnect(nil)(c)

	// Verify callback is set to nil
	if c.onConnectCallback != nil {
		t.Error("expected onConnectCallback to be nil")
	}
}

func TestWithOnNewMessage(t *testing.T) {
	c := &client{}
	called := false
	var receivedData []byte
	callback := func(data []byte) {
		called = true
		receivedData = data
	}

	// Apply WithOnNewMessage option
	WithOnNewMessage(callback)(c)

	// Verify callback is set
	if c.onNewMessage == nil {
		t.Error("expected onNewMessage to be set")
	}

	// Test that the callback works
	testData := []byte("test message")
	c.onNewMessage(testData)
	if !called {
		t.Error("expected callback to be called")
	}
	if string(receivedData) != string(testData) {
		t.Errorf("expected received data to be %s, got %s", string(testData), string(receivedData))
	}
}

func TestWithOnNewMessageNil(t *testing.T) {
	c := &client{}

	// Apply WithOnNewMessage option with nil
	WithOnNewMessage(nil)(c)

	// Verify callback is set to nil
	if c.onNewMessage != nil {
		t.Error("expected onNewMessage to be nil")
	}
}

func TestMultipleOptions(t *testing.T) {
	c := &client{}
	customHost := "wss://test.example.com"
	customInterval := 15 * time.Second
	customLogger := NewLogger()

	// Apply multiple options
	WithHost(customHost)(c)
	WithPingInterval(customInterval)(c)
	WithLogger(customLogger)(c)

	// Verify all options are applied correctly
	if c.host != customHost {
		t.Errorf("expected host to be %s, got %s", customHost, c.host)
	}

	if c.pingInterval != customInterval {
		t.Errorf("expected ping interval to be %v, got %v", customInterval, c.pingInterval)
	}

	if c.logger != customLogger {
		t.Error("expected logger to be set to custom logger")
	}

}

func TestOptionChaining(t *testing.T) {
	c := &client{}
	customHost := "wss://chained.example.com"
	customInterval := 20 * time.Second

	// Chain multiple options
	WithHost(customHost)(c)
	WithPingInterval(customInterval)(c)

	// Verify chained options work correctly
	if c.host != customHost {
		t.Errorf("expected host to be %s, got %s", customHost, c.host)
	}

	if c.pingInterval != customInterval {
		t.Errorf("expected ping interval to be %v, got %v", customInterval, c.pingInterval)
	}
}

func TestConstants(t *testing.T) {
	// Test that constants are properly defined
	if defaultHost == "" {
		t.Error("defaultHost should not be empty")
	}

	if defaultPingInterval <= 0 {
		t.Error("defaultPingInterval should be positive")
	}

	// Verify default values are reasonable
	if defaultPingInterval != 5*time.Second {
		t.Errorf("expected defaultPingInterval to be 5 seconds, got %v", defaultPingInterval)
	}
}
