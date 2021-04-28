package upbit

import (
	"fmt"
	"github.com/jinzhu/configor"
)

type config struct {
	Bot struct {
		AccessKey string `required:"true"`
		SecretKey string `required:"true"`
	}
}

func NewConfig(filename string) (*config, error) {
	conf := config{}

	err := configor.New(&configor.Config{Silent: true}).Load(&conf, filename)
	if (config{}) == conf {
		return nil, fmt.Errorf("Failed to find configuration `%s`", filename)
	}
	if err != nil {
		return nil, err
	}

	return &conf, nil
}
