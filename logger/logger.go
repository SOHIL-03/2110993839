package logger

import (
	"fmt"
	"github.com/SOHIL-03/2110993839/config"
	"gopkg.in/natefinch/lumberjack.v2"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.Logger

func Init() *zap.Logger {
	environment := config.GetString("environment")

	var err error
	if environment != "development" {
		writer := zapcore.AddSync(&lumberjack.Logger{
			Filename:   "service.log",
			MaxSize:    10, // megabytes
			MaxBackups: 5,
			MaxAge:     10, // days
		})
		location, _ := time.LoadLocation("Asia/Kolkata")
		cfg := zap.NewProductionEncoderConfig()
		cfg.EncodeTime = func(t time.Time, pae zapcore.PrimitiveArrayEncoder) {
			t = t.In(location)
			zapcore.ISO8601TimeEncoder(t, pae)
		}
		cfg.EncodeDuration = zapcore.MillisDurationEncoder
		core := zapcore.NewTee(zapcore.NewCore(
			zapcore.NewJSONEncoder(cfg),
			writer,
			zap.DebugLevel,
		))
		logger = zap.New(core)
	} else {
		config := zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		logger, err = config.Build()
	}
	if err != nil {
		panic(fmt.Errorf("unable to initialize logger\n %w", err))
	}
	defer logger.Sync()

	logger = logger.WithOptions(zap.AddCaller())
	return logger
}

func GetLogger() *zap.Logger {
	return logger
}
