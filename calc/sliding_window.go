package calc

import (
	"errors"

	"github.com/ha2398/vwap/feed"
)

// slidingWindow that maintains the partial VWAP calculation data for each
// update, in a queue fashion.
//
// Since VWAP is SUM_i(price_i * size_i) / SUM_i(size_i), to calculate the
// VWAP on each update, we have the following:
// 		1) If the sliding window has the desired amount of elements: in this
// case a new update means the oldest data must be removed from the
// calculation, while the new data must be incorporated.
//		2) If the sliding window does not have the desired amount of elements:
// in this case we simply add the new data to the result.
//
// In both these cases, the update data is added to the queue, so that the
// sliding window keeps moving when needed.
type slidingWindow struct {
	// Queue of partial VWAP calculation data.
	data chan vwapPartialData

	// Current length.
	size, currentLength int

	// Indicates if the desired number of elements has been reached.
	isWindowFull bool

	// Current VWAP.
	vwap float64

	// Current sums used in the VWAP calculation.
	// VWAP is vwapNumerator / vwapDenominator
	vwapNumerator, vwapDenominator float64
}

func newSlidingWindow(size int) *slidingWindow {
	return &slidingWindow{
		data: make(chan vwapPartialData, size),
		size: size,
	}
}

func (w *slidingWindow) getVWAP() float64 {
	return w.vwap
}

func (w *slidingWindow) addMatch(match feed.Match) error {
	currentPartial := getVWAPPartialDataFromMatch(match)

	if w.isWindowFull {
		// If we enter this case, the sliding window is full. Hence, drop the
		// oldest entry and remove its data from the partial sums.
		select {
		case oldestPartial := <-w.data:
			// Success receiving from channel.
			w.subtractPartial(oldestPartial)

		default:
			// This should never happen. Error out.
			return errors.New("unable to receive from data channel")
		}
	}

	select {
	case w.data <- currentPartial:
		// Success pushing to channel.
		w.addPartial(currentPartial)

		// Only update the length until the window is full.
		if !w.isWindowFull {
			w.currentLength += 1
			if w.currentLength == w.size {
				w.isWindowFull = true
			}
		}

	default:
		// This should never happen. Error out.
		return errors.New("unable to send to data channel")
	}

	// After updating the partial sums for VWAP calculation, update the VWAP
	// value.
	if w.vwapDenominator == 0 {
		return errors.New("unable to calculate VWAP with zero denominator")
	}

	w.vwap = w.vwapNumerator / w.vwapDenominator
	return nil
}

func (w *slidingWindow) addPartial(partialData vwapPartialData) {
	w.vwapNumerator += partialData.product
	w.vwapDenominator += partialData.size
}

func (w *slidingWindow) subtractPartial(partialData vwapPartialData) {
	w.vwapNumerator -= partialData.product
	w.vwapDenominator -= partialData.size
}

// vwapPartialData holds a pair of values, the product (price_i * size_i) and
// the size_i, for a given update i.
type vwapPartialData struct {
	product, size float64
}

func getVWAPPartialDataFromMatch(match feed.Match) vwapPartialData {
	return vwapPartialData{
		product: match.Price * match.Size,
		size:    match.Size,
	}
}
