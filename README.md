# Go Polymarket real-time data client

A Go client for receiving real-time data messages from Polymarket's WebSocket API. This client provides a simple and efficient way to subscribe to live market data, including trades, orders, comments, and RFQ (Request for Quote) activities.

## About

This Go client was inspired by [Polymarket's official TypeScript real-time data client](https://github.com/Polymarket/real-time-data-client).


## Installation

```bash
go get github.com/Matthew17-21/go-polymarket-real-time-data-client
```

## Quick Start

```go
package main

import (
    "fmt"
    "time"
    
    polymarketdataclient "github.com/Matthew17-21/go-polymarket-real-time-data-client"
)

func main() {
    // Create a new client with options
    client := polymarketdataclient.New(
        polymarketdataclient.WithLogger(polymarketdataclient.NewLogger()),
        polymarketdataclient.WithOnConnect(func() {
            fmt.Println("Connected to Polymarket WebSocket!")
        }),
        polymarketdataclient.WithOnNewMessage(func(message []byte) {
            fmt.Printf("Received: %s\n", string(message))

            // Handle any further message processing. Can use the types in `payload.go` to unmarshal
        }),
    )

    // Connect to the server
    if err := client.Connect(); err != nil {
        panic(err)
    }

    // Subscribe to market data
    subscriptions := []polymarketdataclient.Subscription{
        {
            Topic: polymarketdataclient.TopicActivity,
            Type:  polymarketdataclient.MessageTypeAll,
        },
        {
            Topic: polymarketdataclient.TopicComments,
            Type:  polymarketdataclient.MessageTypeCommentCreated,
        },
    }

    if err := client.Subscribe(subscriptions); err != nil {
        panic(err)
    }

    // Keep the connection alive
    time.Sleep(30 * time.Second)

    // Clean up
    client.Disconnect()
}
```

## API Reference

### Client Creation

#### Available Options

- `WithLogger(logger)` - Set a custom logger
- `WithPingInterval(duration)` - Set ping interval (default: 5s)
- `WithHost(host)` - Set WebSocket host (default: wss://ws-live-data.polymarket.com)
- `WithOnConnect(callback)` - Set connection callback
- `WithOnNewMessage(callback)` - Set message received callback

### Topics

Available topics for subscription:

- `TopicActivity` - Market activity data
- `TopicComments` - Comment-related events
- `TopicRfq` - Request for Quote data

### Message Types

Available message types for filtering. More details can be found on the [official documentaiton](https://github.com/Polymarket/real-time-data-client):

#### Activity Topic
- `MessageTypeAll` - All messages (use `"*"`)
- `MessageTypeTrades` - Trade events
- `MessageTypeOrdersMatched` - Order matching events

#### Comments Topic
- `MessageTypeCommentCreated` - New comments
- `MessageTypeCommentRemoved` - Comment deletions
- `MessageTypeReactionCreated` - Reaction additions
- `MessageTypeReactionRemoved` - Reaction removals

#### RFQ Topic
- `MessageTypeRequestCreated` - New RFQ requests
- `MessageTypeRequestEdited` - RFQ request edits
- `MessageTypeRequestCanceled` - RFQ request cancellations
- `MessageTypeRequestExpired` - RFQ request expirations
- `MessageTypeQuoteCreated` - New quotes
- `MessageTypeQuoteEdited` - Quote edits
- `MessageTypeQuoteCanceled` - Quote cancellations
- `MessageTypeQuoteExpired` - Quote expirations

## Performance Benchmarks

This client uses optimized JSON parsing instead of string searching for better performance. Benchmark results show improvements:

```
goos: darwin
goarch: arm64
pkg: github.com/Matthew17-21/go-polymarket-real-time-data-client/internal/json_utils
cpu: Apple M3 Max

BenchmarkIsJsonFormatVsPayloadCheck/IsJsonFormat-14         	29993344	40.68 ns/op
BenchmarkIsJsonFormatVsPayloadCheck/StringContainsPayload-14	15817960	76.14 ns/op
BenchmarkMessageProcessingCheck/IsJsonFormat-14             	62086358	19.02 ns/op
BenchmarkMessageProcessingCheck/StringContainsPayload-14    	29597839	40.87 ns/op
BenchmarkEdgeCases/IsJsonFormat_LongStrings-14              	73811518	15.33 ns/op
BenchmarkEdgeCases/StringContainsPayload_LongStrings-14     	 2879878	415.4 ns/op
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Disclaimer

This is an unofficial client library and is not affiliated with Polymarket. Use at your own risk. Additionally, AI was used to generate *most* tests.

## TODOs

- [ ] Add support for message parsing based on type & topic using `json.RawMessage`