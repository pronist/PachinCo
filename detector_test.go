package upbit

import (
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	"time"
)

var d *detector

func TestNewDetector(t *testing.T) {
	var err error

	d, err = newDetector()
	assert.NoError(t, err)

	assert.NotNil(t, d.ws)
	assert.IsType(t, &websocket.Conn{}, d.ws)
}

func TestDetector_Run(t *testing.T) {
	go d.run(
		&Bot{QuotationClient: &QuotationClient{Client: http.DefaultClient}},
		market,
		func(b *Bot, t map[string]interface{}) bool {
			return true
		})

	timer := time.NewTimer(time.Second * 1)

	select {
	case <-timer.C:
	case tick := <-d.d:
		assert.Equal(t, "ticker", tick["type"].(string))
		return
	}

	assert.Fail(t, "Cannot receive tick from Detector.")
}
