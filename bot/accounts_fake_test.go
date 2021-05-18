package bot

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/boltdb/bolt"
	"github.com/pronist/upbit/utils"
	"github.com/stretchr/testify/assert"
)

var (
	facc    *FakeAccounts
	tmpfile *os.File
)

func TestNewFakeAccounts(t *testing.T) {
	var err error

	tmpfile, err = ioutil.TempFile("", "accounts")
	assert.NoError(t, err)

	facc, err = NewFakeAccounts(tmpfile.Name(), 55000.0)
	assert.NoError(t, err)

	if _, err := os.Stat(tmpfile.Name()); os.IsNotExist(err) {
		assert.Error(t, err)
	}

	err = facc.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(faccAccountsBucketName))
		if err != nil {
			return err
		}

		var krwAccount Account

		encodedKrwAccount := bkt.Get([]byte("KRW"))
		assert.NotNil(t, encodedKrwAccount)

		err := utils.Deserialize(encodedKrwAccount, &krwAccount)
		if err != nil {
			return err
		}

		assert.Equal(t, 55000.0, krwAccount.Balance)

		return nil
	})
	assert.NoError(t, err)
}

func TestFakeAccounts_Accounts(t *testing.T) {
	acc, err := facc.accounts()
	assert.NoError(t, err)

	assert.Contains(t, acc[0], "currency")
	assert.Contains(t, acc[0], "balance")
	assert.Contains(t, acc[0], "avg_buy_price")
}

func TestFakeAccounts_Order(t *testing.T) {
	coin, err := newCoin(facc, "BTC", 0.1)
	assert.NoError(t, err)

	ok, err := facc.order(&Bot{}, coin, b, 2.0, coin.onceOrderPrice)
	assert.NoError(t, err)
	assert.True(t, ok)

	acc, err := facc.accounts()
	assert.NoError(t, err)

	balances := getBalances(acc)

	assert.Equal(t, 55000.0-coin.onceOrderPrice*2.0, balances["KRW"])
	assert.Equal(t, 2.0, balances["BTC"])

	err = facc.db.Close()
	assert.NoError(t, err)

	err = tmpfile.Close()
	assert.NoError(t, err)

	err = os.Remove(tmpfile.Name())
	assert.NoError(t, err)
}
