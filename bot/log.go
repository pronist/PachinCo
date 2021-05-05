package bot

import (
	"github.com/mattn/go-colorable"
	"github.com/sirupsen/logrus"
	"github.com/snowzach/rotatefilehook"
	"time"
)

const LogFileName = "logs/log.log"

var logger *logrus.Logger

type Log struct {
	Msg    interface{}
	Fields logrus.Fields
	Level  logrus.Level
}

func init() {
	rotateFileHook, err := rotatefilehook.NewRotateFileHook(rotatefilehook.RotateFileConfig{
		Filename:   LogFileName,
		MaxSize:    10, // megabytes
		MaxBackups: 3,
		MaxAge:     28, //days
		Formatter: &logrus.JSONFormatter{
			TimestampFormat: time.RFC822,
		},
	})
	if err != nil {
		logrus.Panic(err)
	}

	logger = &logrus.Logger{
		Out: colorable.NewColorableStdout(),
		Formatter: &logrus.TextFormatter{
			ForceColors: true, FullTimestamp: true, TimestampFormat: time.RFC822,
		},
		Hooks: map[logrus.Level][]logrus.Hook{
			logrus.WarnLevel: {rotateFileHook}, logrus.ErrorLevel: {rotateFileHook}, logrus.FatalLevel: {rotateFileHook},
		},
		Level: logrus.InfoLevel,
	}
}
