package log

import (
	"github.com/sirupsen/logrus"
)

type Log struct {
	Msg    interface{}
	Fields logrus.Fields
	Level  logrus.Level
}
