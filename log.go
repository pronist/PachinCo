package upbit

import (
	"github.com/mattn/go-colorable"
	"github.com/sirupsen/logrus"
	"github.com/snowzach/rotatefilehook"
	"time"
)

type Log struct {
	Msg       string
	Fields    logrus.Fields
	Terminate bool
}

func NewLogger(filename string, level logrus.Level) (*logrus.Logger, error) {
	rotateFileHook, err := rotatefilehook.NewRotateFileHook(rotatefilehook.RotateFileConfig{
		Filename:   filename,
		MaxSize:    50, // megabytes
		MaxBackups: 3,
		MaxAge:     28, //days
		Level:      level,
		Formatter: &logrus.JSONFormatter{
			TimestampFormat: time.RFC822,
		},
	})
	if err != nil {
		return nil, err
	}

	logger := &logrus.Logger{
		Out: colorable.NewColorableStdout(),
		Formatter: &logrus.TextFormatter{
			ForceColors: true, FullTimestamp: true, TimestampFormat: time.RFC822,
		},
		Hooks: map[logrus.Level][]logrus.Hook{level: {rotateFileHook}},
		Level: level,
	}

	return logger, nil
}