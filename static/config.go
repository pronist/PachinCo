package static

import (
	"os"
	"path/filepath"
	"time"

	"github.com/jinzhu/configor"
	"github.com/sirupsen/logrus"
)

var Config = struct {
	AccessKey string        `required:"true"` // 엑세스 키
	SecretKey string        `required:"true"` // 비밀 키
	C         float64       `required:"true"` // 비중
	R         float64       `required:"true"` // 주문 가격 비중
	K         float64       `required:"true"` // 변동성 돌파 상수
	Timeout   time.Duration `required:"true"` // 주문 대기 중인 경우 최대 대기시간
	Max       int           `required:"true"` // 최대 추적할 코인의 갯수 (분산투자 비율하고도 관련있음)
}{}

func init() {
	config := ".upbit.yml"

	if env := os.Getenv("APP_ENV"); env == "test" {
		config = filepath.Join("..", config)
	}

	err := configor.New(&configor.Config{Silent: true}).Load(&Config, config)
	if err != nil {
		logrus.Fatal(err)
	}
}
