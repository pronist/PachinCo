package log

import (
	"os"

	"github.com/mattn/go-colorable"
	"github.com/sirupsen/logrus"
	"github.com/snowzach/rotatefilehook"
)

var Logger = make(chan Log) // 외부에서 사용하게 될 로그 채널이다.

func init() {
	go func() {
		timestampFormat := "2006-01-02 15:04:05"
		log := "log.log"

		rotateFileHook, err := rotatefilehook.NewRotateFileHook(rotatefilehook.RotateFileConfig{
			Filename:   log,
			MaxSize:    10, // megabytes
			MaxBackups: 3,
			MaxAge:     28, // days
			Formatter: &logrus.JSONFormatter{
				TimestampFormat: timestampFormat,
			},
		})
		if err != nil {
			logrus.Panic(err)
		}

		logger := &logrus.Logger{
			Out: colorable.NewColorableStdout(),
			Formatter: &logrus.TextFormatter{
				ForceColors: true, FullTimestamp: true, TimestampFormat: timestampFormat,
			},
			Hooks: map[logrus.Level][]logrus.Hook{
				logrus.InfoLevel: {rotateFileHook}, logrus.WarnLevel: {rotateFileHook}, logrus.ErrorLevel: {rotateFileHook}, logrus.FatalLevel: {rotateFileHook},
			},
			Level: logrus.TraceLevel,
		}

		for log := range Logger {
			if env := os.Getenv("APP_ENV"); env != "test" {
				logger.WithFields(log.Fields).Log(log.Level, log.Msg)
			}
		}
	}()
}
