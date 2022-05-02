// +build unit

package feed

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	ws "github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

// testEchoServerHandler spins up a test server that will simply echo all the
// messages it receives back to the sender.
func testEchoServerHandler(w http.ResponseWriter, r *http.Request) {
	wsUpgrader := ws.Upgrader{}
	c, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	defer c.Close()
	for {
		msgType, message, err := c.ReadMessage()
		if err != nil {
			break
		}

		err = c.WriteMessage(msgType, message)
		if err != nil {
			break
		}
	}
}

func Test_CreateSubscription(t *testing.T) {
	// Spin up test server.
	echoServer := httptest.NewServer(http.HandlerFunc(testEchoServerHandler))
	defer echoServer.Close()
	echoEndpoint := strings.Replace(echoServer.URL, "http", "ws", 1)

	testCases := []struct {
		desc          string
		endpoint      string
		productIDs    []string
		expectedError error
	}{
		{
			desc:          "no products, empty endpoint",
			endpoint:      "",
			productIDs:    []string{},
			expectedError: errors.New("error dialing WebSocket endpoint \"\": malformed ws or wss URL"),
		},
		{
			desc:          "no products, valid endpoint",
			endpoint:      echoEndpoint,
			productIDs:    []string{},
			expectedError: nil,
		},
		{
			desc:          "valid endpoint and products",
			endpoint:      echoEndpoint,
			productIDs:    []string{"product1", "product2"},
			expectedError: nil,
		},
	}

	for _, tc := range testCases {
		conn, err := CreateSubscription(tc.endpoint, tc.productIDs)

		assert.Equal(t, tc.expectedError, err,
			"For test %q, got unexpected error value", tc.desc)
		if err != nil {
			continue
		}

		var message Message
		err = conn.ReadJSON(&message)
		assert.Nil(t, err,
			"For test %q, got unexpected error reading from WebSocket connection",
			tc.desc)

		assert.Equal(t, SubscribeType, message[TypeKey],
			"For test %q, got unexpected message type from echo server", tc.desc)
	}
}

func Test_ReadMessages(t *testing.T) {
	// Spin up test server.
	testServerHandler := func(w http.ResponseWriter, r *http.Request) {
		wsUpgrader := ws.Upgrader{}
		c, err := wsUpgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Errorf("Error upgrading HTTP connection to WebSocket: %v", err)
			return
		}

		defer c.Close()

		testMessages := []Message{
			Message{
				"myKey": "he",
			},
			Message{
				"myKey": "ll",
			},
			Message{
				"myKey": "o ",
			},
			Message{
				"myKey": "wo",
			},
			Message{
				"myKey": "rl",
			},
			Message{
				"myKey": "d",
			},
		}
		for _, msg := range testMessages {
			err = c.WriteJSON(msg)
			if err != nil {
				t.Errorf("Error writing message in test server: %v", err)
			}
		}
	}

	testServer := httptest.NewServer(http.HandlerFunc(testServerHandler))
	defer testServer.Close()
	endpoint := strings.Replace(testServer.URL, "http", "ws", 1)

	c, _, err := ws.DefaultDialer.Dial(endpoint, nil)
	if err != nil {
		t.Fatalf("Error dialing test WebSocket server: %v", err)
		return
	}

	// output will be built using the individual messages received from the server.
	var output string
	ReadMessages(c, func(msg Message, err error) {
		if err != nil {
			return
		}

		output += msg["myKey"].(string)
	})

	assert.Equal(t, "hello world", output, "Got wrong result")
}

func Test_ReadMessagesError(t *testing.T) {
	// Spin up test server.
	testServerHandler := func(w http.ResponseWriter, r *http.Request) {
		wsUpgrader := ws.Upgrader{}
		c, err := wsUpgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Errorf("Error upgrading HTTP connection to WebSocket: %v", err)
			return
		}

		defer c.Close()

		msg := Message{
			TypeKey: ErrorType,
		}
		err = c.WriteJSON(msg)
		if err != nil {
			t.Errorf("Error writing message in test server: %v", err)
		}
	}

	testServer := httptest.NewServer(http.HandlerFunc(testServerHandler))
	defer testServer.Close()
	endpoint := strings.Replace(testServer.URL, "http", "ws", 1)

	c, _, err := ws.DefaultDialer.Dial(endpoint, nil)
	if err != nil {
		t.Fatalf("Error dialing test WebSocket server: %v", err)
		return
	}

	// output will be built using the individual messages received from the server.
	ReadMessages(c, func(msg Message, err error) {
		assert.NotNil(t, err)
	})
}
