package upbit

import (
	"github.com/gorilla/websocket"
	uuid "github.com/satori/go.uuid"
)

const (
	sockURL     = "wss://api.upbit.com/websocket"
	sockVersion = "v1"
)

type websocketClient struct {
	ws *websocket.Conn
	data []map[string]interface{}
}

func newWebsocketClient(t string, codes []string, isOnlySnapshot, isOnlyRealtime bool) (*websocketClient, error) {
	ws, _, err := websocket.DefaultDialer.Dial(sockURL+"/"+sockVersion, nil)
	if err != nil {
		return nil, err
	}

	// https://docs.upbit.com/docs/upbit-quotation-websocket
	data := []map[string]interface{}{
		{"ticket": uuid.NewV4()}, // ticket
		{"type": t, "codes": codes, "isOnlySnapshot": isOnlySnapshot, "isOnlyRealtime": isOnlyRealtime}, // type
		// format
	}

	return &websocketClient{ws, data}, nil
}