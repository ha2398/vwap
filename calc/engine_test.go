package calc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
