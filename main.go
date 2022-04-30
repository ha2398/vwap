package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/ha2398/vwap/calc"
	"github.com/ha2398/vwap/feed"
)

// createInterruptChannel creates and returns a channel that notifies on
// interrupt signals.
func createInterruptChannel() chan os.Signal {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	return interrupt
}

func main() {
	// Create channel to detect interrupt signals.
	interruptCh := createInterruptChannel()

	// Create connection to feed and subscribe to channels of interest.
	// TODO: Receive these arguments as CLI parameters.
	webSocketFeedEndpoint := "wss://ws-feed.exchange.coinbase.com"
	subscriptionChannels := []string{"matches"}
	tradingPairs := []string{"BTC-USD"}
	feedConn, err := feed.CreateSubscription(webSocketFeedEndpoint,
		subscriptionChannels, tradingPairs)
	if err != nil {
		log.Fatalf("Error creating feed subscription: %v\n", err)
		return
	}
	defer feedConn.Close()

	// Create calculation engine.
	vwapEngine, err := calc.NewEngine(feedConn)
	if err != nil {
		log.Fatalf("Error creating new VWAP calculation engine: %v\n", err)
		return
	}

	// Start reading messages.
	doneCh := vwapEngine.Run()

	// Block until we are either done reading messages, or an interrupt signal
	// is detected.
	select {
	case <-doneCh:
	case <-interruptCh:
	}
}
