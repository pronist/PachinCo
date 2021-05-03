package upbit

import (
	"github.com/jinzhu/configor"
	"github.com/sirupsen/logrus"
	"time"
)

const ConfigFileName = "upbit.config.yml"

var Config = struct {
	Coins   map[string]float64 `required:"true"` // 트래킹할 코인 목록
	R       float64            `required:"true"` // 비중
	Timeout time.Duration      `required:"true"` // 주문이 대기 중인 경우 최대 대기시간
}{}

func init() {
	err := configor.New(&configor.Config{Silent: true}).Load(&Config, ConfigFileName)

	if err != nil {
		logrus.Fatal(err)
	}
}
