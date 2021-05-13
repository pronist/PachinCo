package accounts

import (
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/fatih/structs"
	"github.com/pronist/upbit"
	"github.com/pronist/upbit/bot"
	"github.com/pronist/upbit/log"
	"github.com/pronist/upbit/utils"
	"github.com/sirupsen/logrus"
	"strconv"
)

const (
	dbName = "test_accounts.db"
	accountsBucketName = "accounts"
)

// 하나의 계정을 나타낸다.
// upbit.Accounts 에 대해서는 별도로 변환하지 않기 때문에 테스트용도로만 쓴다.
type Account struct {
	Currency      string  `structs:"currency"` // 기준 단위
	Balance       float64 `structs:"balance"` // 수량
	AvgBuyPrice   float64 `structs:"avg_buy_price"` // 매수 평균가
}

// TestAccounts 는 테스트용 계정을 나타내며, 이는 dbName 에 있는 파일에 영구적으로 저장 될 것이다.
// 계정 정보가 업비트 서버에 저장되어 있는 것을 시뮬레이트 하기 위함이다.
type TestAccounts struct {
	db *bolt.DB // 계정 정보가 들어있는 데이터베이스
}

func NewTestAccounts(krw float64) (*TestAccounts, error) {
	db, err := bolt.Open(dbName, 0666, nil)
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		// 버킷이 없을 때만 새로운 자금 할당 처리
		// accountsBucketName 에 해당하는 버킷이 없다는 것은 데이터베이스를 완전히 새로 만들었다는 것을 의미한다.
		if bkt := tx.Bucket([]byte(accountsBucketName)); bkt == nil {
			bkt, err := tx.CreateBucket([]byte(accountsBucketName))
			if err != nil {
				return err
			}

			// 초기 자금을 할당한다.
			account := &Account{Currency: "KRW", Balance: krw, AvgBuyPrice: 0}

			encodedAccount, err := utils.Serialize(account)
			if err != nil {
				return err
			}
			if err := bkt.Put([]byte("KRW"), encodedAccount); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	//
	log.Logger <- log.Log{
		Msg: "Creating new accounts for Testing.",
		Fields: logrus.Fields{
			"KRW": krw,
		},
		Level: logrus.DebugLevel,
	}
	//

	return &TestAccounts{db}, nil
}

func (acc *TestAccounts) Order(coin *bot.Coin, side string, volume, price float64, t map[string]interface{}) (bool, error) {
	err := acc.db.Update(func(tx *bolt.Tx) error {
		var krwAccount, account Account

		bkt := tx.Bucket([]byte(accountsBucketName))
		if bkt == nil {
			return fmt.Errorf("'%#v' bucket not found", accountsBucketName)
		}

		// 매수/매도를 나타내기 위한 부호
		// 매수 -> 1, 매도 -> -1
		var sign float64

		sign = 1
		if side == upbit.S {
			sign = -sign
		}

		// KRW 계정에서 잔고를 매수/매도에 따라 증가/감소 시킨다.
		encodedKrwAccount := bkt.Get([]byte("KRW"))

		err := utils.Deserialize(encodedKrwAccount, &krwAccount)
		if err != nil {
			return err
		}
		krwAccount.Balance += -sign * volume * price

		encodedKrwAccount, err = utils.Serialize(krwAccount)
		if err != nil {
			return err
		}
		if err := bkt.Put([]byte("KRW"), encodedKrwAccount); err != nil {
			return err
		}

		// 코인에 대한 계정을 얻는다.
		encodedAccount := bkt.Get([]byte(coin.Name))

		// 만약 계정이 아직 없다면 새로 만든다.
		if encodedAccount == nil {
			newAccount := &Account{Currency: coin.Name, Balance: 0, AvgBuyPrice: 0}

			encodedNewAccount, err := utils.Serialize(newAccount)
			if err != nil {
				return err
			}
			if err := bkt.Put([]byte(coin.Name), encodedNewAccount); err != nil {
				return err
			}

			encodedAccount = encodedNewAccount
		}
		err = utils.Deserialize(encodedAccount, &account)
		if err != nil {
			return err
		}

		// 매수평균가 및 잔액
		p := account.AvgBuyPrice * account.Balance + (sign * volume * price) // 총 매수 가격
		b := account.Balance + (sign * volume) // 총 매수 수량

		account.AvgBuyPrice = p / b
		account.Balance += sign * volume

		encodedAccount, err = utils.Serialize(account)
		if err != nil {
			return err
		}
		if err := bkt.Put([]byte(coin.Name), encodedAccount); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return false, err
	}

	err = coin.Refresh(acc)
	if err != nil {
		return false, err
	}

	//
	log.Logger <- log.Log{
		Msg: "ORDER",
		Fields: logrus.Fields{
			"coin": coin.Name, "side": side, "volume": volume, "price": price, "change-rate": t["change_rate"].(float64),
		},
		Level: logrus.WarnLevel,
	}
	//

	return true, nil
}

func (acc *TestAccounts) Accounts() ([]map[string]interface{}, error) {
	r := make([]map[string]interface{}, 0)

	err := acc.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(accountsBucketName))
		if bkt == nil {
			return fmt.Errorf("'%#v' bucket not found", accountsBucketName)
		}

		return bkt.ForEach(func(coin, encodedAccount []byte) error {
			var account Account

			err := utils.Deserialize(encodedAccount, &account)
			if err != nil {
				return err
			}

			m := structs.Map(account)

			for k, v := range m {
				// 기존의 업비트 계정에서 사용하는 표현 방식과 호환성을 위해 문자열로 변환
				if v, ok := v.(float64); ok {
					m[k] = strconv.FormatFloat(v, 'g', 8, 64)
				}
			}
			r = append(r, m)

			return nil
		})
	})
	if err != nil {
		return nil, err
	}

	return r, nil
}
