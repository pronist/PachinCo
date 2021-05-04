package bot

import (
	"github.com/pronist/upbit"
	"github.com/sirupsen/logrus"
	"time"
)

func (c *Coin) Tick() {
	for {
		ticker, err := upbit.API.GetTicker("KRW-" + c.Name)
		if err != nil {
			LogChan <- Log{Msg: err, Level: logrus.ErrorLevel}
		}

		// 실행 중인 전략의 수 만큼 보내면 코인에 적용된 모든 전략이 틱을 수신할 수 있다.
		// 전략은 반드시 시작할 때 틱을 소비해야 한다.
		for range c.Strategies {
			c.Ticker <- ticker
		}

		time.Sleep(time.Second * 1)
	}
}
