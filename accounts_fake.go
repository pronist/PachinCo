package upbit

import (
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/fatih/structs"
	"github.com/sirupsen/logrus"
	"strconv"
)

const (
	FaccDbName             = "accounts.db"
	faccAccountsBucketName = "accounts"
)

// 하나의 계정을 나타낸다.
// upbit.Accounts 에 대해서는 별도로 변환하지 않기 때문에 테스트용도로만 쓴다.
type Account struct {
	Currency    string  `structs:"currency"`      // 기준 단위
	Balance     float64 `structs:"balance"`       // 수량
	AvgBuyPrice float64 `structs:"avg_buy_price"` // 매수 평균가
}

// FakeAccounts 는 테스트용 계정을 나타내며, 이는 dbName 에 있는 파일에 영구적으로 저장 될 것이다.
// 계정 정보가 업비트 서버에 저장되어 있는 것을 시뮬레이트 하기 위함이다.
// 또한 FakeAccounts 는 Accounts 인터페이스를 따른다.
type FakeAccounts struct {
	db *bolt.DB // 계정 정보가 들어있는 데이터베이스
}

// NewFakeAccounts 는 새로운 FakeAccounts 를 만든다.
// faccDbName 에 해당하는 boltDB 를 만들고 krw 만큼의 자금을 faccAccountsBucketName 에 할당한다.
func NewFakeAccounts(dbname string, krw float64) (*FakeAccounts, error) {
	db, err := bolt.Open(dbname, 0666, nil)
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		// 버킷이 없을 때만 새로운 자금 할당 처리
		// accountsBucketName 에 해당하는 버킷이 없다는 것은 데이터베이스를 완전히 새로 만들었다는 것을 의미한다.
		if bkt := tx.Bucket([]byte(faccAccountsBucketName)); bkt == nil {
			bkt, err := tx.CreateBucket([]byte(faccAccountsBucketName))
			if err != nil {
				return err
			}

			// 초기 자금을 할당한다.
			account := &Account{Currency: "KRW", Balance: krw, AvgBuyPrice: 0}

			encodedAccount, err := Serialize(account)
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
	Logger <- Log{
		Msg: "Creating new accounts for Testing.",
		Fields: logrus.Fields{
			"KRW": krw,
		},
		Level: logrus.DebugLevel,
	}
	//

	return &FakeAccounts{db}, nil
}

// Order 는 주문을 처리한다. 실제 업비트 서버에 보내는 것이 아니므로 이를 모조한다.
// 주문을 보낸 뒤에는 faccAccountsBucketName 의 내용이 변경되며
func (acc *FakeAccounts) Order(_ *Bot, coin *Coin, side string, volume, price float64) (bool, error) {
	a, err := acc.Accounts()
	if err != nil {
		return false, err
	}

	balances := GetBalances(a)

	if GetAverageBuyPrice(a, coin.Name)*balances[coin.Name]+coin.OnceOrderPrice <= coin.Limit {
		err := acc.db.Update(func(tx *bolt.Tx) error {
			var krwAccount, account Account

			bkt := tx.Bucket([]byte(faccAccountsBucketName))
			if bkt == nil {
				return fmt.Errorf("'%#v' bucket not found", faccAccountsBucketName)
			}

			// 매수/매도를 나타내기 위한 부호
			// 매수: 1, 매도: -1
			var sign float64

			sign = 1
			if side == S {
				sign = -sign
			}

			// KRW 계정에서 잔고를 매수/매도에 따라 증가/감소 시킨다.
			encodedKrwAccount := bkt.Get([]byte("KRW"))

			err := Deserialize(encodedKrwAccount, &krwAccount)
			if err != nil {
				return err
			}
			krwAccount.Balance += -sign * volume * price

			encodedKrwAccount, err = Serialize(krwAccount)
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

				encodedNewAccount, err := Serialize(newAccount)
				if err != nil {
					return err
				}
				if err := bkt.Put([]byte(coin.Name), encodedNewAccount); err != nil {
					return err
				}

				encodedAccount = encodedNewAccount
			}
			err = Deserialize(encodedAccount, &account)
			if err != nil {
				return err
			}

			// 매수평균가 및 잔액
			p := account.AvgBuyPrice*account.Balance + (sign * volume * price)
			b := account.Balance + (sign * volume)

			account.AvgBuyPrice = p / b
			account.Balance += sign * volume

			encodedAccount, err = Serialize(account)
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
		Logger <- Log{
			Msg: "ORDER",
			Fields: logrus.Fields{
				"coin": coin.Name, "side": side, "volume": volume, "price": price,
			},
			Level: logrus.WarnLevel,
		}
		//
	}

	return true, nil
}

// Accounts 는 accounts_utils 에 정의된 함수들과 호환,
// UpbitAccounts 가 가지고 있는 구조와 동일하게 하기 위해 구조를 맞춘다.
func (acc *FakeAccounts) Accounts() ([]map[string]interface{}, error) {
	r := make([]map[string]interface{}, 0)

	err := acc.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(faccAccountsBucketName))
		if bkt == nil {
			return fmt.Errorf("'%#v' bucket not found", faccAccountsBucketName)
		}
		return bkt.ForEach(func(coin, encodedAccount []byte) error {
			var account Account

			err := Deserialize(encodedAccount, &account)
			if err != nil {
				return err
			}

			m := structs.Map(account)

			for k, v := range m {
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
