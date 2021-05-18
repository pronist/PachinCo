package bot

import (
	"github.com/pronist/upbit/client"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	"time"
)

func TestDetector_Run(t *testing.T) {
	d, err := newDetector()
	assert.NoError(t, err)

	go d.run(
		&Bot{QuotationClient: &client.QuotationClient{Client: http.DefaultClient}},
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
