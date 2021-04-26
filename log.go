package upbit

import (
	"github.com/mattn/go-colorable"
	"github.com/sirupsen/logrus"
	"github.com/snowzach/rotatefilehook"
	"os"
	"time"
)

type Log struct {
	msg       string
	fields    logrus.Fields
	terminate bool
}

// Loggers
const (
	errLog = "logs/error.log"
	logLog = "logs/log.log"
)

var (
	errLogger = logrus.New() // 에러 로깅
	logLogger = logrus.New() // 매수/매도 등 이벤트 로깅
	stdLogger = logrus.New() // 가격 tick 로깅
)

func setConsoleAndFileLog(logger *logrus.Logger, filename string, loglevel logrus.Level) {
	rotateFileHook, err := rotatefilehook.NewRotateFileHook(rotatefilehook.RotateFileConfig{
		Filename:   filename,
		MaxSize:    50, // megabytes
		MaxBackups: 3,
		MaxAge:     28, //days
		Level:      loglevel,
		Formatter: &logrus.JSONFormatter{
			TimestampFormat: time.RFC822,
		},
	})
	if err != nil {
		logrus.Fatalf("Failed to initialize file rotate hook: %v", err)
	}

	logger.SetFormatter(&logrus.TextFormatter{
		ForceColors: true, FullTimestamp: true, TimestampFormat: time.RFC822,
	})

	logger.SetLevel(loglevel)
	logger.SetOutput(colorable.NewColorableStdout())

	logger.AddHook(rotateFileHook)
}

// 초기화 작업, 주로 로거 초기화.
func init() {
	setConsoleAndFileLog(errLogger, errLog, logrus.ErrorLevel)
	setConsoleAndFileLog(logLogger, logLog, logrus.WarnLevel)

	// stdLogger
	stdLogger.SetFormatter(&logrus.TextFormatter{
		ForceColors: true, FullTimestamp: true, TimestampFormat: time.RFC822,
	})
	stdLogger.SetOutput(os.Stdout)
}
