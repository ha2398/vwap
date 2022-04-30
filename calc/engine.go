// Package calc provides an engine for calculating the VWAP for trading pairs.
package calc

import (
	"errors"
	"log"

	ws "github.com/gorilla/websocket"
	"github.com/ha2398/vwap/feed"
)

// Size of the buffered channel to use for passing match data from the reader
// goroutine to the handler one.
const bufferedChannelSize int = 1000

// Engine is the calculator engine for VWAP.
type Engine struct {
	feedConn *ws.Conn
}

func NewEngine(feedConn *ws.Conn) (*Engine, error) {
	if feedConn == nil {
		return nil, errors.New("unable to create calculation engine with " +
			"nil feed connection")
	}

	return &Engine{
		feedConn: feedConn,
	}, nil
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
			log.Printf("Error parsing match data: %v\n", err)
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
		// TODO
		log.Printf("Match: %+v\n", match)
	}
}
