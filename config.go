package upbit

import (
	"fmt"
	"github.com/jinzhu/configor"
	"reflect"
)

type Config struct {
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
