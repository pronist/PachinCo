package upbit

import (
	"github.com/boltdb/bolt"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

const dbName = "test_accounts.db"

var facc *FakeAccounts

func TestNewFakeAccounts(t *testing.T) {
	var err error

	if _, err := os.Stat(dbName); os.IsExist(err) {
		err := os.Remove(dbName)
		assert.NoError(t, err)
	}

	facc, err = NewFakeAccounts(dbName, 55000.0)
	assert.NoError(t, err)

	if _, err := os.Stat(dbName); os.IsNotExist(err) {
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

		err := Deserialize(encodedKrwAccount, &krwAccount)
		if err != nil {
			return err
		}

		assert.Equal(t, 55000.0, krwAccount.Balance)

		return nil
	})
	assert.NoError(t, err)
}

func TestFakeAccounts_Accounts(t *testing.T) {
	acc, err := facc.Accounts()
	assert.NoError(t, err)

	assert.Contains(t, acc[0], "currency")
	assert.Contains(t, acc[0], "balance")
	assert.Contains(t, acc[0], "avg_buy_price")
}

func TestFakeAccounts_Order(t *testing.T) {
	coin, err := NewCoin(facc, "BTC", 0.1)
	assert.NoError(t, err)

	ok, err := facc.Order(&Bot{}, coin, B, 2.0, coin.OnceOrderPrice)
	assert.NoError(t, err)
	assert.True(t, ok)

	acc, err := facc.Accounts()
	assert.NoError(t, err)

	balances := GetBalances(acc)

	assert.Equal(t, 55000.0-coin.OnceOrderPrice*2.0, balances["KRW"])
	assert.Equal(t, 2.0, balances["BTC"])

	err = facc.db.Close()
	assert.NoError(t, err)

	err = os.Remove(dbName)
	assert.NoError(t, err)
}
