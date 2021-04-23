package upbit

import (
	"fmt"
	"time"
)

type Bot struct {
	*Client
	*QuotationClient
}

func (b *Bot) Watch(market string) {
	for {
		day, _ := b.QuotationClient.Get("/candles/days", map[string]string{
			"market": market, "count": "1",
		})
		if day, ok := day.([]interface{}); ok {
			if candle, ok := day[0].(map[string]interface{}); ok {
				fmt.Printf("\r%s: %.4f%%", market, candle["change_rate"].(float64))
			}
		}

		time.Sleep(1 * time.Second)
	}
}
