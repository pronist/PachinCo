package upbit

import (
	"github.com/mattn/go-colorable"
	"github.com/sirupsen/logrus"
	"github.com/snowzach/rotatefilehook"
	"time"
)

var (
	ErrLogger = NewLogger("logs/error.log", logrus.ErrorLevel, true)
	LogLogger = NewLogger("logs/log.log", logrus.WarnLevel, false)
	StdLogger = NewLogger("", logrus.InfoLevel, false)
)

func NewLogger(filename string, level logrus.Level, caller bool) *logrus.Logger {
	rotateFileHook, err := rotatefilehook.NewRotateFileHook(rotatefilehook.RotateFileConfig{
		Filename:   filename,
		MaxSize:    10, // megabytes
		MaxBackups: 3,
		MaxAge:     28, //days
		Level:      level,
		Formatter: &logrus.JSONFormatter{
			TimestampFormat: time.RFC822,
		},
	})
	if err != nil {
		Exit <- err.Error()
	}

	logger := &logrus.Logger{
		Out: colorable.NewColorableStdout(),
		Formatter: &logrus.TextFormatter{
			ForceColors: true, FullTimestamp: true, TimestampFormat: time.RFC822,
		},
		Hooks:        map[logrus.Level][]logrus.Hook{level: {rotateFileHook}},
		Level:        level,
		ReportCaller: caller,
	}

	return logger
}
