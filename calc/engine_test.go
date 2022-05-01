// +build unit

package calc

import (
	"errors"
	"testing"

	ws "github.com/gorilla/websocket"
	"github.com/ha2398/vwap/feed"
	"github.com/stretchr/testify/assert"
)

func Test_NewEngine(t *testing.T) {
	testCases := []struct {
		desc          string
		conn          *ws.Conn
		tradingPairs  []string
		windowSize    int
		expectedError error
	}{
		{
			desc:          "nil feed connection",
			conn:          nil,
			tradingPairs:  []string{},
			windowSize:    0,
			expectedError: errors.New("nil feed connection"),
		},
		{
			desc:          "no trading pairs",
			conn:          &ws.Conn{},
			tradingPairs:  []string{},
			windowSize:    0,
			expectedError: errors.New("no trading pairs"),
		},
		{
			desc:          "invalid window size",
			conn:          &ws.Conn{},
			tradingPairs:  []string{"myPair"},
			windowSize:    -42,
			expectedError: errors.New("invalid window size -42, must be at least 1"),
		},
		{
			desc:          "valid parameters",
			conn:          &ws.Conn{},
			tradingPairs:  []string{"myPair"},
			windowSize:    42,
			expectedError: nil,
		},
	}

	for _, tc := range testCases {
		engine, err := NewEngine(tc.conn, tc.tradingPairs, tc.windowSize)
		assert.Equal(t, tc.expectedError, err,
			"For test %q, got unexpected error value", tc.desc)

		if err != nil {
			continue
		}

		assert.NotNil(t, engine.feedConn,
			"For test %q, got nil feed connection", tc.desc)
		assert.Equal(t, tc.tradingPairs, engine.tradingPairs,
			"For test %q, got incorrect trading pairs", tc.desc)
		assert.Equal(t, len(tc.tradingPairs), len(engine.vwapValues),
			"For test %q, vwapValues slice not initialized correctly", tc.desc)
		assert.Equal(t, tc.windowSize, engine.windowSize,
			"For test %q, got incorrect window size", tc.desc)
	}
}

func Test_getWindowForProduct(t *testing.T) {
	testCases := []struct {
		desc      string
		engine    *Engine
		productID string
	}{
		{
			desc: "no previous window",
			engine: &Engine{
				windowSize: 42,
				windows:    map[string]*slidingWindow{},
			},
			productID: "someID",
		},
		{
			desc: "previous window present",
			engine: &Engine{
				windowSize: 42,
				windows: map[string]*slidingWindow{
					"someID": newSlidingWindow(42),
				},
			},
			productID: "someID",
		},
	}

	for _, tc := range testCases {
		output := tc.engine.getWindowForProduct(tc.productID)

		assert.NotNil(t, output,
			"For test %q, got nil output", tc.desc)
		assert.NotNil(t, output.data,
			"For test %q, got nil data channel", tc.desc)
		assert.Equal(t, tc.engine.windowSize, output.size,
			"For test %q, got incorrect window size", tc.desc)
	}
}

func Test_handleMatches(t *testing.T) {
	testCases := []struct {
		desc         string
		tradingPairs []string
		matches      []feed.Match
		expectedLog  string
	}{
		{
			desc:         "no matches",
			tradingPairs: []string{"pair1"},
			matches:      []feed.Match{},
			expectedLog:  "\"pair1\": 0.000000",
		},
		{
			desc:         "only last_match data",
			tradingPairs: []string{"pair1", "pair2"},
			matches: []feed.Match{
				feed.Match{
					IsLast:    true,
					Price:     10,
					ProductID: "pair2",
					Size:      2,
				},
			},
			expectedLog: "\"pair1\": 0.000000, \"pair2\": 10.000000",
		},
		{
			desc:         "last_match and match data",
			tradingPairs: []string{"pair1", "pair2"},
			matches: []feed.Match{
				feed.Match{
					IsLast:    true,
					Price:     10,
					ProductID: "pair2",
					Size:      2,
				},
				feed.Match{
					IsLast:    true,
					Price:     5,
					ProductID: "pair1",
					Size:      1,
				},
				feed.Match{
					IsLast:    false,
					Price:     20,
					ProductID: "pair1",
					Size:      10,
				},
			},
			expectedLog: "\"pair1\": 18.636364, \"pair2\": 10.000000",
		},
		{
			desc:         "invalid match data",
			tradingPairs: []string{"pair1"},
			matches: []feed.Match{
				feed.Match{
					IsLast:    false,
					Price:     0,
					ProductID: "pair1",
					Size:      0,
				},
			},
			expectedLog: "\"pair1\": 0.000000",
		},
	}

	for _, tc := range testCases {
		engine, err := NewEngine(&ws.Conn{}, tc.tradingPairs, 10)

		assert.Nil(t, err, "For test %q, got error creating engine", tc.desc)

		matchCh := make(chan feed.Match, bufferedChannelSize)
		doneCh := make(chan struct{})

		for _, match := range tc.matches {
			matchCh <- match
		}

		close(matchCh)
		engine.handleMatches(matchCh, doneCh)

		assert.Equal(t, tc.expectedLog, engine.getVWAPLog(),
			"For test %q, for unexpected VWAP log", tc.desc)
	}
}

func Test_getVWAPLogFormat(t *testing.T) {
	testCases := []struct {
		desc           string
		tradingPairs   []string
		expectedOutput string
	}{
		{
			desc:           "nil input",
			tradingPairs:   nil,
			expectedOutput: "",
		},
		{
			desc:           "no trading pairs",
			tradingPairs:   []string{},
			expectedOutput: "",
		},
		{
			desc:           "single trading pair",
			tradingPairs:   []string{"pair1"},
			expectedOutput: `"pair1": %f`,
		},
		{
			desc:           "multiple trading pairs",
			tradingPairs:   []string{"pair1", "pair2", "pair3"},
			expectedOutput: `"pair1": %f, "pair2": %f, "pair3": %f`,
		},
	}

	for _, tc := range testCases {
		output := getVWAPLogFormat(tc.tradingPairs)
		assert.Equal(t, tc.expectedOutput, output,
			"For test %q, got wrong output", tc.desc)
	}
}

func Test_getVWAPLog(t *testing.T) {
	testCases := []struct {
		desc           string
		tradingPairs   []string
		windows        map[string]*slidingWindow
		expectedOutput string
	}{
		{
			desc:           "nil input",
			tradingPairs:   nil,
			windows:        nil,
			expectedOutput: "",
		},
		{
			desc:           "no trading pairs",
			tradingPairs:   []string{},
			windows:        map[string]*slidingWindow{},
			expectedOutput: "",
		},
		{
			desc:         "single trading pair",
			tradingPairs: []string{"pair1"},
			windows: map[string]*slidingWindow{
				"pair1": &slidingWindow{vwap: 10.0},
			},
			expectedOutput: "\"pair1\": 10.000000",
		},
		{
			desc:         "multiple trading pairs",
			tradingPairs: []string{"pair1", "pair2", "pair3"},
			windows: map[string]*slidingWindow{
				"pair1": &slidingWindow{vwap: 10.0},
				"pair2": &slidingWindow{vwap: 42.123456},
				"pair3": &slidingWindow{vwap: -123.456789},
			},
			expectedOutput: "\"pair1\": 10.000000, \"pair2\": 42.123456, \"pair3\": -123.456789",
		},
	}

	for _, tc := range testCases {
		e := &Engine{
			tradingPairs:  tc.tradingPairs,
			vwapValues:    make([]interface{}, len(tc.tradingPairs)),
			vwapLogFormat: getVWAPLogFormat(tc.tradingPairs),
			windows:       tc.windows,
		}
		output := e.getVWAPLog()
		assert.Equal(t, tc.expectedOutput, output,
			"For test %q, got wrong output", tc.desc)
	}
}
