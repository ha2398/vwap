package feed

import (
	"fmt"
	"log"
	"strconv"

	ws "github.com/gorilla/websocket"
)

// Keys used in messages.
const (
	ChannelsKey   string = "channels"
	MessageKey    string = "message"
	PriceKey      string = "price"
	ProductIDKey  string = "product_id"
	ProductIDsKey string = "product_ids"
	SideKey       string = "side"
	SizeKey       string = "size"
	TypeKey       string = "type"
)

// Message types.
const (
	ErrorType     string = "error"
	MatchType     string = "match"
	LastMatchType string = "last_match"
	SubscribeType string = "subscribe"
	UnknownType   string = "unknown"
)

//
// Messages.
//

type Message map[string]interface{}

func (m *Message) GetValueForKey(key string) string {
	if m == nil {
		return ""
	}

	rawValue, hasKey := (*m)[key]
	if !hasKey {
		return ""
	}

	value, ok := rawValue.(string)
	if !ok {
		return ""
	}

	return value
}

func newSubscribeMessage(channels, productIDs []string) Message {
	return Message{
		TypeKey:       SubscribeType,
		ChannelsKey:   channels,
		ProductIDsKey: productIDs,
	}
}

// Match represents the relevant data contained in a match for VWAP
// calculation.
type Match struct {
	Price     float64
	ProductID string
	Side      string
	Size      float64
}

// ParseMatch tries and parses a Match from the given message passed as
// argument. It returns the parsed match, a bool indicating if the given
// message contains a match at all, and any error found in the parsing process.
func ParseMatch(msg Message) (Match, bool, error) {
	msgType := msg.GetValueForKey(TypeKey)
	if msgType != MatchType && msgType != LastMatchType {
		return Match{}, false, nil
	}

	priceStr := msg.GetValueForKey(PriceKey)
	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		return Match{}, true, fmt.Errorf("error parsing %q field: %v",
			PriceKey, err)
	}

	sizeStr := msg.GetValueForKey(SizeKey)
	size, err := strconv.ParseFloat(sizeStr, 64)
	if err != nil {
		return Match{}, true, fmt.Errorf("error parsing %q field: %v",
			SizeKey, err)
	}

	return Match{
		Price:     price,
		ProductID: msg.GetValueForKey(ProductIDKey),
		Side:      msg.GetValueForKey(SideKey),
		Size:      size,
	}, true, nil
}

//
// Message methods.
//

// ReadMessages takes a WebSocket connection and reads incoming messages from
// it. For each message received, it calls the messageCallback function.
func ReadMessages(conn *ws.Conn, messageCallback func(Message, error)) {
	for {
		var message Message
		err := conn.ReadJSON(&message)
		messageCallback(message, err)

		if err != nil {
			log.Fatalf("Error reading JSON WebSocket message: %v\n", err)
			return
		}
	}
}
