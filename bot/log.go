package bot

import (
	"github.com/mattn/go-colorable"
	"github.com/sirupsen/logrus"
	"github.com/snowzach/rotatefilehook"
	"time"
)

type Log struct {
	Msg    interface{}
	Fields logrus.Fields
	Level  logrus.Level
}

var Logger = make(chan Log) // 외부에서 사용하게 될 로그 채널이다.

func init() {
	go func() {
		rotateFileHook, err := rotatefilehook.NewRotateFileHook(rotatefilehook.RotateFileConfig{
			Filename:   "logs/log.log",
			MaxSize:    10, // megabytes
			MaxBackups: 3,
			MaxAge:     28, //days
			Formatter: &logrus.JSONFormatter{
				TimestampFormat: time.RFC822,
			},
		})
		if err != nil {
			logrus.Fatal(err)
		}

		logger := &logrus.Logger{
			Out: colorable.NewColorableStdout(),
			Formatter: &logrus.TextFormatter{
				ForceColors: true, FullTimestamp: true, TimestampFormat: time.RFC822,
			},
			Hooks: map[logrus.Level][]logrus.Hook{
				// 로그 레벨이 Warn, Error, Fatal 인 경우 파일에 기록한다.
				logrus.WarnLevel: {rotateFileHook}, logrus.ErrorLevel: {rotateFileHook}, logrus.FatalLevel: {rotateFileHook},
			},
			Level: logrus.InfoLevel,
		}

		for {
			select{
			// 지속적으로 로그를 받아온다. 이 시점에서 가장 깔끔한 로그 처리방법인 듯보인다.
			case log := <-Logger:
				logger.WithFields(log.Fields).Log(log.Level, log.Msg)
			}
		}
	}()
}
