package api

import (
	"github.com/pronist/upbit/gateway"
	"time"
)

func (api *API) WaitUntilCompletedOrder(uuid string) error {
	for {
		order, err := api.Client.CallWith("GET", "/order", gateway.Query{"uuid": uuid})
		if err != nil {
			return err
		}
		if order, ok := order.(map[string]interface{}); ok {
			if order["state"].(string) == "done" {
				return nil
			}
		}

		time.Sleep(1 * time.Second)
	}
}