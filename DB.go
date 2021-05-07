package upbit

import (
	"github.com/boltdb/bolt"
	"github.com/sirupsen/logrus"
)

var Db *bolt.DB

const (
	DbName          = "upbit.db"
	CoinsBucketName = "coins"
)

func init() {
	var err error

	if Db, err = bolt.Open(DbName, 0666, nil); err == nil {
		err = Db.Update(func(tx *bolt.Tx) error {
			var err error

			// 초기화를 위해 이미 버킷이 존재할 경우 버킷을 날린다.
			if b := tx.Bucket([]byte(CoinsBucketName)); b != nil {
				if err := tx.DeleteBucket([]byte(CoinsBucketName)); err != nil {
					return err
				}
			}

			// `coins` 버킷은 현재 추적 중인 코인들을 담아둔다.
			// 실행 중인 모든 전략은 `coins` 버킷에
			// 트래킹 중인 코인이 담겨있는지 검사하고 없으면 전략을 중단한다.
			if _, err = tx.CreateBucketIfNotExists([]byte(CoinsBucketName)); err != nil {
				return err
			}

			return nil
		})
		if err != nil {
			logrus.Fatal(err)
		}
	}
}
