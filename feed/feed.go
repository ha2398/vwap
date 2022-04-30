// Package feed provides the object model and operations for sending and
// receiving WebSocket messages through the Coinbase exchange feed.
package feed

import (
	"fmt"

	ws "github.com/gorilla/websocket"
)

// CreateSubscription takes as arguments the WebSocket endpoint to connect to,
// the channels to subscribe to, and product IDs of interest. It creates and
// returns a connection to the given endpoint, and subscribes to the given
// channels/products.
func CreateSubscription(
	endpoint string, channels, productIDs []string,
) (*ws.Conn, error) {
	// Connect to WebSocket endpoint.
	c, _, err := ws.DefaultDialer.Dial(endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("error dialing WebSocket endpoint %q: %v",
			endpoint, err)
	}

	// Subscribe to channels.
	subscribeMessage := newSubscribeMessage(channels, productIDs)
	if err := c.WriteJSON(subscribeMessage); err != nil {
		c.Close()
		return nil, fmt.Errorf("error writing subscribe message to WebSocket "+
			"endpoint: %v", err)
	}

	return c, nil
}
