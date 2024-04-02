package logging

import (
	kzap "github.com/go-kratos/kratos/contrib/log/zap/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/qiulin/kratos-boot/sharedconf"
	slogzap "github.com/samber/slog-zap/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log/slog"
)

func NewZapLogger(c *sharedconf.Logging) (*zap.Logger, error) {
	cl := "DEBUG"
	if c != nil {
		cl = c.Level
	}
	zapDev := true
	if c != nil && c.Zap != nil {
		zapDev = !c.Zap.Production
	}

	var config zap.Config
	if zapDev {
		config = zap.NewDevelopmentConfig()
	} else {
		config = zap.NewProductionConfig()
	}
	config.DisableCaller = true
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.Sampling = nil
	level, err := zapcore.ParseLevel(cl)
	if err != nil {
		level = zapcore.DebugLevel
	}
	config.Level = zap.NewAtomicLevelAt(level)
	return config.Build()
}

func NewLogger(zlog *zap.Logger) log.Logger {
	z := kzap.NewLogger(zlog)
	log.SetLogger(z)
	return z
}

func NewSLogger(level string, zlog *zap.Logger) *slog.Logger {
	l := slog.LevelDebug
	_ = l.UnmarshalText([]byte(level))
	logger := slog.New(slogzap.Option{Level: l, Logger: zlog}.NewZapHandler())
	slog.SetDefault(logger)
	return logger
}
