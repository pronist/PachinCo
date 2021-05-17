package upbit

import (
	"github.com/sirupsen/logrus"
)

type log struct {
	msg    interface{}
	fields logrus.Fields
	level  logrus.Level
}
