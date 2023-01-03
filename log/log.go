package log

import (
	"fmt"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.SugaredLogger
var initialized = false

func MustSync() {
	if err := logger.Sync(); err != nil {
		// see: https://github.com/uber-go/zap/issues/880
	}
}

type LoggerOption struct {
	Level string
}

func InitLogger(opt *LoggerOption) error {
	loggerConfig := zap.NewProductionConfig()
	lv, err := zap.ParseAtomicLevel(opt.Level)
	var keeperr error
	if err != nil {
		keeperr = err
		lv = zap.NewAtomicLevelAt(zap.InfoLevel)
	}
	loggerConfig.Level = lv
	loggerConfig.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339)
	log, err := loggerConfig.Build()
	if err != nil {
		return err
	}
	if keeperr != nil {
		log.Error(fmt.Sprintf("invalid log level: %s. use `info` instead", opt.Level))
	}
	zap.ReplaceGlobals(log)
	logger = log.Sugar()
	initialized = true
	return err
}

func GetLogger() *zap.SugaredLogger {
	if !initialized {
		if err := InitLogger(&LoggerOption{}); err != nil {
			panic(err)
		}
	}
	return logger
}
