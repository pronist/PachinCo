package client

import (
	"github.com/gorilla/websocket"
	uuid "github.com/satori/go.uuid"
)

const (
	sockURL     = "wss://api.upbit.com/websocket"
	sockVersion = "v1"
)

type WebsocketClient struct {
	Ws   *websocket.Conn
	Data []map[string]interface{}
}

func NewWebsocketClient(t string, codes []string, isOnlySnapshot, isOnlyRealtime bool) (*WebsocketClient, error) {
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

	return &WebsocketClient{ws, data}, nil
}
