package upbit

import (
	"github.com/jinzhu/configor"
	"github.com/sirupsen/logrus"
	"time"
)

const ConfigFileName = "upbit.config.yml"

var Config = struct {
	C       float64       `required:"true"` // 비중
	R       float64       `required:"true"` // 주문 가격 비중
	K       float64       `required:"true"` // 변동성 돌파 상수
	Timeout time.Duration `required:"true"` // 주문 대기 중인 경우 최대 대기시간
}{}

func init() {
	err := configor.New(&configor.Config{Silent: true}).Load(&Config, ConfigFileName)

	if err != nil {
		logrus.Fatal(err)
	}
}
