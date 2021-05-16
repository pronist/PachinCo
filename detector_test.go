package upbit

import (
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	"time"
)

var detector *Detector

func TestNewDetector(t *testing.T) {
	var err error

	detector, err = NewDetector()
	assert.NoError(t, err)

	assert.NotNil(t, detector.ws)
	assert.IsType(t, &websocket.Conn{}, detector.ws)
}

func TestDetector_Run(t *testing.T) {
	go detector.Run(
		&Bot{QuotationClient: &QuotationClient{Client: http.DefaultClient}},
		Market,
		func(b *Bot, t map[string]interface{}) bool {
			return true
		})

	timer := time.NewTimer(time.Second * 1)

	select {
	case <-timer.C:
		assert.Fail(t, "Cannot receive tick from Detector.")
	case tick := <-detector.D:
		assert.Equal(t, "ticker", tick["type"].(string))
	}
}
