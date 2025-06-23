package main

import (
	"fmt"
	"time"

	polymarketdataclient "github.com/Matthew17-21/go-polymarket-real-time-data-client"
)

func main() {
	fmt.Println("Starting the example client...")
	client := polymarketdataclient.New(
		polymarketdataclient.WithLogger(polymarketdataclient.NewLogger()),
		polymarketdataclient.WithOnConnect(func() {
			fmt.Println("Connected to the server!")
		}),
		polymarketdataclient.WithOnNewMessage(func(message []byte) {
			fmt.Println("[onNewMessage] Received message:", string(message))
		}),
	)
	err := client.Connect()
	if err != nil {
		fmt.Println("Error connecting to the server:", err)
		return
	}

	err = client.Subscribe([]polymarketdataclient.Subscription{
		{Topic: polymarketdataclient.TopicActivity, Type: polymarketdataclient.MessageTypeAll},
		{Topic: polymarketdataclient.TopicComments, Type: polymarketdataclient.MessageTypeAll},
		{Topic: polymarketdataclient.TopicRfq, Type: polymarketdataclient.MessageTypeAll},
	})
	if err != nil {
		fmt.Println("Error subscribing to topics:", err)
		return
	}

	time.Sleep(5 * time.Second)
	err = client.Disconnect()
	if err != nil {
		fmt.Println("Error disconnecting from the server:", err)
		return
	}

	time.Sleep(30 * time.Second)
}
