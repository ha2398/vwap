// +build unit

package calc

import (
	"errors"
	"testing"

	"github.com/ha2398/vwap/feed"
	"github.com/stretchr/testify/assert"
)

func Test_newSlidingWindow(t *testing.T) {
	testCases := []struct {
		desc         string
		size         int
		expectedSize int
	}{
		{
			desc:         "negative size",
			size:         -10,
			expectedSize: 1,
		},
		{
			desc:         "zero size",
			size:         0,
			expectedSize: 1,
		},
		{
			desc:         "valid size",
			size:         42,
			expectedSize: 42,
		},
	}

	for _, tc := range testCases {
		window := newSlidingWindow(tc.size)

		assert.Equal(t, tc.expectedSize, window.size,
			"For test %q, got unexpected window size", tc.desc)
		assert.NotNil(t, window.data,
			"For test %q, got nil window data", tc.desc)
	}
}

func Test_addMatch(t *testing.T) {
	testCases := []struct {
		desc                    string
		window                  *slidingWindow
		enqueuedPartialData     []vwapPartialData
		match                   feed.Match
		expectedError           error
		expectWindowFull        bool
		expectedLength          int
		expectedVWAPNumerator   float64
		expectedVWAPDenominator float64
		expectedVWAP            float64
	}{
		{
			desc: "window full, but no data available",
			window: &slidingWindow{
				isWindowFull: true,
			},
			expectedError: errors.New("unable to receive from data channel"),
		},
		{
			desc: "window full",
			window: &slidingWindow{
				isWindowFull:  true,
				size:          100,
				currentLength: 100,
			},
			enqueuedPartialData: []vwapPartialData{
				vwapPartialData{
					product: 2,
					size:    1,
				},
			},
			match: feed.Match{
				Price: 10,
				Size:  2,
			},
			expectedError:           nil,
			expectWindowFull:        true,
			expectedLength:          100,
			expectedVWAPNumerator:   18,
			expectedVWAPDenominator: 1,
			expectedVWAP:            18,
		},
		{
			desc: "zero VWAP denominator",
			window: &slidingWindow{
				isWindowFull:  true,
				size:          100,
				currentLength: 100,
			},
			enqueuedPartialData: []vwapPartialData{
				vwapPartialData{
					product: 0,
					size:    0,
				},
			},
			match: feed.Match{
				Price: 0,
				Size:  0,
			},
			expectedError: errors.New("unable to calculate VWAP with zero denominator"),
		},
		{
			desc: "window not full",
			window: &slidingWindow{
				isWindowFull:  false,
				size:          3,
				currentLength: 2,
			},
			enqueuedPartialData: []vwapPartialData{
				vwapPartialData{
					product: 2,
					size:    1,
				},
				vwapPartialData{
					product: 2,
					size:    1,
				},
			},
			match: feed.Match{
				Price: 10,
				Size:  2,
			},
			expectedError:           nil,
			expectWindowFull:        true,
			expectedLength:          3,
			expectedVWAPNumerator:   20,
			expectedVWAPDenominator: 2,
			expectedVWAP:            10,
		},
	}

	for _, tc := range testCases {
		tc.window.data = make(chan vwapPartialData, tc.window.size)
		for _, data := range tc.enqueuedPartialData {
			tc.window.data <- data
		}

		err := tc.window.addMatch(tc.match)

		assert.Equal(t, tc.expectedError, err,
			"For test %q, got unexpected error value", tc.desc)

		if err != nil {
			continue
		}

		assert.Equal(t, tc.expectWindowFull, tc.window.isWindowFull,
			"For test %q, got unexpected isWindowFull value", tc.desc)
		assert.Equal(t, tc.expectedLength, tc.window.currentLength,
			"For test %q, got unexpected length value", tc.desc)
		assert.Equal(t, tc.expectedVWAPNumerator, tc.window.vwapNumerator,
			"For test %q, got unexpected VWAP numerator", tc.desc)
		assert.Equal(t, tc.expectedVWAPDenominator, tc.window.vwapDenominator,
			"For test %q, got unexpected VWAP denominator", tc.desc)
		assert.Equal(t, tc.expectedVWAP, tc.window.vwap,
			"For test %q, got unexpected VWAP", tc.desc)
	}
}
