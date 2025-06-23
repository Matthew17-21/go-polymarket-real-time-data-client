package polymarketdataclient

import "time"

const (
	defaultHost         = "wss://ws-live-data.polymarket.com"
	defaultPingInterval = 5 * time.Second
)

type ClientOptions func(*client)

// defaultOpts sets the default options for the client
func defaultOpts() ClientOptions {
	return func(c *client) {
		c.host = defaultHost
		c.pingInterval = defaultPingInterval
		c.logger = NewSilentLogger()
	}
}

// WithLogger sets the logger for the client
func WithLogger(l Logger) ClientOptions {
	return func(c *client) {
		c.logger = l
	}
}

// WithPingInterval sets the ping interval for the client
func WithPingInterval(interval time.Duration) ClientOptions {
	return func(c *client) {
		c.pingInterval = interval
	}
}

// WithHost sets the host for the client
func WithHost(host string) ClientOptions {
	return func(c *client) {
		c.host = host
	}
}

// WithOnConnect sets the onConnect callback for the client
func WithOnConnect(f func()) ClientOptions {
	return func(c *client) {
		c.onConnectCallback = f
	}
}

func WithOnNewMessage(f func([]byte)) ClientOptions {
	return func(c *client) {
		c.onNewMessage = f
	}
}
