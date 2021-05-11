package log

import (
	"github.com/mattn/go-colorable"
	"github.com/sirupsen/logrus"
	"github.com/snowzach/rotatefilehook"
)

var Logger = make(chan Log) // 외부에서 사용하게 될 로그 채널이다.

func init() {
	go func() {
		timestampFormat := "2006-01-02 15:04:05"

		rotateFileHook, err := rotatefilehook.NewRotateFileHook(rotatefilehook.RotateFileConfig{
			Filename:   "log.log",
			MaxSize:    10, // megabytes
			MaxBackups: 3,
			MaxAge:     28, //days
			Formatter: &logrus.JSONFormatter{
				TimestampFormat: timestampFormat,
			},
		})
		if err != nil {
			logrus.Fatal(err)
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

		for {
			select {
			// 지속적으로 로그를 받아온다. 이 시점에서 가장 깔끔한 로그 처리방법인 듯보인다.
			case log := <-Logger:
				logger.WithFields(log.Fields).Log(log.Level, log.Msg)
			}
		}
	}()
}
