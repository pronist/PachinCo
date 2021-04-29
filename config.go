package upbit

import (
	"fmt"
	"github.com/jinzhu/configor"
	"reflect"
)

type config struct {
	KeyPair struct {
		AccessKey string `required:"true"`
		SecretKey string `required:"true"`
	}
	Coins map[string]float64 `required:"true"`
	F     float64            `required:"true"`
	L     float64            `required:"true"`
	H     float64            `required:"true"`
	R     float64            `required:"true"`
}

func newConfig(filename string) (*config, error) {
	conf := config{}

	err := configor.New(&configor.Config{Silent: true}).Load(&conf, filename)
	if reflect.DeepEqual(conf, config{}) {
		return nil, fmt.Errorf("Failed to find configuration `%s`", filename)
	}
	if err != nil {
		return nil, err
	}

	return &conf, nil
}
