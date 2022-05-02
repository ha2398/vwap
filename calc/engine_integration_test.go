// +build integration

package calc

import (
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	ws "github.com/gorilla/websocket"
	"github.com/ha2398/vwap/feed"
	"github.com/stretchr/testify/assert"
)

var engineTestMessages []feed.Message = []feed.Message{
	feed.Message{ // Last match for A-B
		"maker_order_id": "id1",
		"price":          "200.0",
		"product_id":     "A-B",
		"sequence":       "12394.123784912",
		"side":           "buy",
		"size":           "5.0",
		"taker_order_id": "id2",
		"time":           "2022-05-01T18:09:24.450429Z",
		"trade_id":       "id3",
		"type":           "last_match",
	},
	feed.Message{ // Last match for C-D
		"maker_order_id": "id4",
		"price":          "10.0",
		"product_id":     "C-D",
		"sequence":       "12394.123784912",
		"side":           "sell",
		"size":           "10.0",
		"taker_order_id": "id8",
		"time":           "2022-05-01T18:09:24.450429Z",
		"trade_id":       "id23",
		"type":           "last_match",
	},
	feed.Message{ // Match for C-D
		"maker_order_id": "id4",
		"price":          "12.0",
		"product_id":     "C-D",
		"sequence":       "12394.123784912",
		"side":           "sell",
		"size":           "1.0",
		"taker_order_id": "id8",
		"time":           "2022-05-01T18:09:24.450429Z",
		"trade_id":       "id23",
		"type":           "match",
	},
	feed.Message{ // Message with non interesting type
		"type": "randomType",
	},
	feed.Message{ // Malformed match message for A-B
		"maker_order_id": "id4123",
		"price":          "hello world",
		"product_id":     "A-B",
		"sequence":       "12394.1284912",
		"side":           "buy",
		"size":           "2.0",
		"taker_order_id": "id8123",
		"time":           "2022-05-01T18:09:24.450429Z",
		"trade_id":       "id23123",
		"type":           "match",
	},
	feed.Message{ // Match for A-B
		"maker_order_id": "id4123",
		"price":          "10",
		"product_id":     "A-B",
		"sequence":       "12394.1284912",
		"side":           "buy",
		"size":           "3",
		"taker_order_id": "id8123",
		"time":           "2022-05-01T18:09:24.450429Z",
		"trade_id":       "id23123",
		"type":           "match",
	},
	feed.Message{ // Match for C-D
		"maker_order_id": "id4123",
		"price":          "21",
		"product_id":     "C-D",
		"sequence":       "12394.1284912",
		"side":           "sell",
		"size":           "10",
		"taker_order_id": "id8123",
		"time":           "2022-05-01T18:09:24.450429Z",
		"trade_id":       "id23123",
		"type":           "match",
	},
	feed.Message{ // Match for C-D, oldest data removed from window
		"maker_order_id": "id4123",
		"price":          "14",
		"product_id":     "C-D",
		"sequence":       "12394.1284912",
		"side":           "sell",
		"size":           "4",
		"taker_order_id": "id8123",
		"time":           "2022-05-01T18:09:24.450429Z",
		"trade_id":       "id23123",
		"type":           "match",
	},
}

// testServerHandler spins up a test server that will simply echo all the
// messages it receives back to the sender.
func testServerHandler(w http.ResponseWriter, r *http.Request) {
	wsUpgrader := ws.Upgrader{}
	c, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	defer c.Close()

	var msg feed.Message
	err = c.ReadJSON(&msg)
	if err != nil {
		log.Printf("Error reading subscription request in test server: %v",
			err)
		return
	}

	err = c.WriteJSON(&msg)
	if err != nil {
		log.Printf("Error writing subscription response in test server: %v",
			err)
		return
	}

	for _, msg := range engineTestMessages {
		err := c.WriteJSON(msg)
		if err != nil {
			log.Printf("Error writing message in test server: %v", err)
			return
		}
	}
}

func Test_EngineRun(t *testing.T) {
	var expectedVWAPFinalValues []interface{} = []interface{}{
		128.75,             // A-B
		18.533333333333335, // C-D
	}

	// Spin up test servers.
	server := httptest.NewServer(http.HandlerFunc(testServerHandler))
	defer server.Close()
	serverEndpoint := strings.Replace(server.URL, "http", "ws", 1)
	tradingPairs := []string{"A-B", "C-D"}

	feedConn, err := feed.CreateSubscription(serverEndpoint, tradingPairs)
	if err != nil {
		t.Fatalf("Error creating feed subscription: %v", err)
		return
	}

	defer feedConn.Close()

	// Create calculation engine.
	windowSize := 3
	vwapEngine, err := NewEngine(feedConn, tradingPairs, windowSize)
	if err != nil {
		t.Fatalf("Error creating new VWAP calculation engine: %v", err)
		return
	}

	// Start calculation engine.
	doneCh := vwapEngine.Run()
	<-doneCh

	// Get VWAP values.
	assert.EqualValues(t, expectedVWAPFinalValues, vwapEngine.vwapValues,
		"Got incorrect VWAP values")
}
