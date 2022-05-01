// Package calc provides an engine for calculating the VWAP for trading pairs.
package calc

import (
	"errors"
	"fmt"
	"log"

	ws "github.com/gorilla/websocket"
	"github.com/ha2398/vwap/feed"
)

// Size of the buffered channel to use for passing match data from the reader
// goroutine to the handler one.
const bufferedChannelSize int = 1000

// Engine is the calculator engine for VWAP.
type Engine struct {
	// Connection to the WebSocket feed.
	feedConn *ws.Conn

	// Trading pairs to calculate VWAP for.
	tradingPairs []string

	// VWAP values for each trading pair, in the same order as they appear in
	// the field tradingPairs.
	vwapValues []interface{}

	// Format string to use for printing VWAPs.
	vwapLogFormat string

	// Sliding window size.
	windowSize int

	// Sliding windows with calculation data for each trading pair.
	windows map[string]*slidingWindow
}

// NewEngine creates a new VWAP calculation engine, using the given connection
// to the WebSocket feed, the trading pairs to calculate VWAP for, and the size
// of the sliding window to use for the algorithm.
func NewEngine(
	feedConn *ws.Conn, tradingPairs []string, windowSize int,
) (*Engine, error) {
	// Sanity checks.
	if feedConn == nil {
		return nil, errors.New("nil feed connection")
	}

	if len(tradingPairs) < 1 {
		return nil, errors.New("no trading pairs")
	}

	if windowSize < 1 {
		return nil, fmt.Errorf("invalid window size %d, must be at least 1",
			windowSize)
	}

	return &Engine{
		feedConn:      feedConn,
		tradingPairs:  tradingPairs,
		vwapValues:    make([]interface{}, len(tradingPairs)),
		vwapLogFormat: getVWAPLogFormat(tradingPairs),
		windows:       make(map[string]*slidingWindow),
		windowSize:    windowSize,
	}, nil
}

// getVWAPLogFormat returns the format string to use when printing VWAPs.
func getVWAPLogFormat(tradingPairs []string) string {
	formatString := ""
	for i, pair := range tradingPairs {
		formatString += fmt.Sprintf("%q: %%f", pair)

		if i != len(tradingPairs)-1 {
			formatString += ", "
		}
	}
	return formatString
}

// getWindowForProduct returns the sliding window for the given product ID. If
// no window is found, one is created and stored in the engine.
func (e *Engine) getWindowForProduct(id string) *slidingWindow {
	window, hasWindow := e.windows[id]
	if !hasWindow {
		window = newSlidingWindow(e.windowSize)
		e.windows[id] = window
	}

	return window
}

// Run is responsible for reading from the WebSocket feed and calculating the
// VWAP for each registered trading pair.
func (e *Engine) Run() chan struct{} {
	// The doneCh is used by the handler goroutine to indicate termination.
	// The main goroutine listens for this event.
	doneCh := make(chan struct{})

	// The matchCh is used to communicate match data between the reader and
	// the handler goroutines.
	matchCh := make(chan feed.Match, bufferedChannelSize)

	// Spin up goroutine to handle incoming matches.
	go e.handleMatches(matchCh, doneCh)

	// Spin up goroutine to read feed messages, parse them, and feed
	// calculation data into the engine.
	go feed.ReadMessages(e.feedConn, func(msg feed.Message, readErr error) {
		if readErr != nil {
			close(matchCh)
			return
		}

		match, isMatch, err := feed.ParseMatch(msg)
		if !isMatch {
			return
		}

		if err != nil {
			log.Printf("Error parsing match data: %v", err)
			return
		}

		matchCh <- match
	})

	return doneCh
}

// handleMatches takes all incoming matches data and updates the VWAP for each
// of them.
// The matchCh argument is used to receive match data, and the doneCh is used
// to communicate the calculation termination.
func (e *Engine) handleMatches(matchCh chan feed.Match, doneCh chan struct{}) {
	defer close(doneCh)
	for match := range matchCh {
		// Get the sliding window for the given trading pair.
		slidingWindow := e.getWindowForProduct(match.ProductID)

		// Update VWAP.
		if err := slidingWindow.addMatch(match); err != nil {
			log.Printf("Error adding match data for %q VWAP calculation: %v",
				match.ProductID, err)
			continue
		}

		// Print current VWAP for each pair.
		// In case the data comes from a last_match message, we don't print
		// it, since not all VWAPs may have been calculated at this point.
		if !match.IsLast {
			log.Print(e.getVWAPLog())
		}
	}
}

// getVWAPLog prints the current VWAP values for all trading pairs of interest.
func (e *Engine) getVWAPLog() string {
	logString := e.vwapLogFormat
	for i, pair := range e.tradingPairs {
		e.vwapValues[i] = e.getWindowForProduct(pair).getVWAP()
	}
	return fmt.Sprintf(logString, e.vwapValues...)
}
