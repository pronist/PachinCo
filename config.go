package upbit

import (
	"github.com/jinzhu/configor"
)

type config struct {
	Bot struct {
		AccessKey string
		SecretKey string
	}
}

func NewConfig(filename string) *config {
	conf := config{}
	configor.Load(&conf, filename)

	return &conf
}
