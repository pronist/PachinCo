package bot

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var c *coin

func TestNewCoin(t *testing.T) {
	var err error

	tmpfile, err = ioutil.TempFile("", "accounts")
	assert.NoError(t, err)

	facc, err = NewFakeAccounts(tmpfile.Name(), 55000.0)
	assert.NoError(t, err)

	c, err = newCoin(facc, "BTC", 0.1)
	assert.NoError(t, err)

	assert.IsType(t, &coin{}, c)
}

func TestCoin_Refresh(t *testing.T) {
	err := c.refresh(facc)
	assert.NoError(t, err)

	acc, err := facc.accounts()
	assert.NoError(t, err)

	assert.Equal(t, getTotalBalance(acc, getBalances(acc))*c.rate, c.limit)
	assert.Equal(t, c.limit*0.5, c.onceOrderPrice)

	err = facc.db.Close()
	assert.NoError(t, err)

	err = tmpfile.Close()
	assert.NoError(t, err)

	err = os.Remove(tmpfile.Name())
	assert.NoError(t, err)
}
