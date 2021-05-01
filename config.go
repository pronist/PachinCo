package upbit

import (
	"fmt"
	"github.com/jinzhu/configor"
	"reflect"
	"time"
)

type Config struct {
	KeyPair struct {
		AccessKey string `required:"true"` // 엑세스키
		SecretKey string `required:"true"` // 비밀키
	}
	Coins map[string]float64 `required:"true"` // 트래킹할 코인 목록
	F     float64            `required:"true"` // 코인을 처음 구매할 때 고려할 하락 기준
	L     float64            `required:"true"` // 구입 하락 기준
	H     float64            `required:"true"` // 판매 상승 기준
	R     float64            `required:"true"` // 비중
	Timeout time.Duration    `required:"true"` // 주문이 대기 중인 경우 최대 대기시간
}

func NewConfig(filename string) (*Config, error) {
	conf := Config{}

	err := configor.New(&configor.Config{Silent: true}).Load(&conf, filename)
	if reflect.DeepEqual(conf, Config{}) {
		return nil, fmt.Errorf("Failed to find configuration `%s`", filename)
	}
	if err != nil {
		return nil, err
	}

	return &conf, nil
}
