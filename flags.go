package main

import (
	"flag"
	"log"
	"strings"
)

// Defaults.
var defaultTradingPairs strSlice = strSlice{"BTC-USD", "ETH-USD", "ETH-BTC"}

const (
	defaultFeedEndpoint string = "wss://ws-feed.exchange.coinbase.com"
	defaultWindowSize   int    = 200
)

// Parameters.
var (
	feedEndpoint string
	tradingPairs strSlice
	windowSize   int
)

// Flag names.
const (
	feedEndpointFlag string = "feed-endpoint"
	tradingPairsFlag string = "trading-pairs"
	windowSizeFlag   string = "window-size"
)

type strSlice []string

func (ss *strSlice) String() string {
	var output string
	numElements := len(*ss)
	for i, s := range *ss {
		output += s

		if i != numElements-1 {
			output += ","
		}
	}
	return output
}

func (ss *strSlice) Set(value string) error {
	*ss = nil

	for _, i := range strings.Split(value, ",") {
		*ss = append(*ss, strings.TrimSpace(i))
	}
	return nil
}

func initFlags() {
	flag.StringVar(&feedEndpoint, feedEndpointFlag, defaultFeedEndpoint,
		"WebSocket endpoint to get match data from")
	flag.Var(&tradingPairs, tradingPairsFlag,
		"comma separated list of trading pairs to calculate VWAP for")
	flag.IntVar(&windowSize, windowSizeFlag, defaultWindowSize,
		"Size of the sliding window to use for VWAP calculation")
	flag.Parse()

	if len(tradingPairs) == 0 {
		tradingPairs = defaultTradingPairs
	}

	// Print values for each parameter.
	log.Printf("WebSocket feed endpoint: %q", feedEndpoint)
	log.Printf("Trading pairs: %v", tradingPairs)
	log.Printf("Window size: %d", windowSize)
}
