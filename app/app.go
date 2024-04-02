package app

import "C"
import (
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/google/wire"
	"github.com/qiulin/kratos-boot/logging"
	"github.com/qiulin/kratos-boot/sharedconf"
	"go.uber.org/zap"
	"log/slog"
)

type Bootstrap struct {
	klogger log.Logger
	zlogger *zap.Logger
	slogger *slog.Logger
	C       config.Config
	Meta    *ServiceMeta
}

func (c *Bootstrap) KLogger() log.Logger {
	return c.klogger
}

func (c *Bootstrap) ZLogger() *zap.Logger {
	return c.zlogger
}

func (c *Bootstrap) SLogger() *slog.Logger {
	return c.slogger
}

func (c *Bootstrap) ScanRootConfig(in interface{}) error {
	return c.C.Scan(in)
}

func (c *Bootstrap) ScanConfig(key string, in interface{}) error {
	return c.C.Value(key).Scan(in)
}

func (c *Bootstrap) Log() *slog.Logger {
	return c.SLogger()
}

type ServiceMeta struct {
	ServiceID   string
	ServiceName string
	Version     string
}

type Options struct {
	ConfigPath, ServiceId, ServiceName, Version string
}

func (opt *Options) EnsureDefaults() {

}

func NewBootstrap(opt *Options) (*Bootstrap, func(), error) {
	opt.EnsureDefaults()
	c := config.New(
		config.WithSource(
			file.NewSource(opt.ConfigPath),
		))
	if err := c.Load(); err != nil {
		return nil, nil, err
	}
	meta := &ServiceMeta{
		ServiceID:   opt.ServiceId,
		ServiceName: opt.ServiceName,
		Version:     opt.Version,
	}

	clog := &sharedconf.Logging{}
	if err := c.Value("logging").Scan(clog); err != nil {
		return nil, nil, err
	}

	zlogger, _ := logging.NewZapLogger(clog)
	zlogger = zlogger.With(zap.String("service_id", opt.ServiceId), zap.String("service_name", opt.ServiceName), zap.String("version", opt.Version))
	logger := logging.NewLogger(zlogger)
	slogger := logging.NewSLogger(zlogger)
	return &Bootstrap{
			klogger: logger,
			zlogger: zlogger,
			slogger: slogger,
			C:       c,
			Meta:    meta,
		}, func() {
			c.Close()
		}, nil
}

func ExportLogger(f *Bootstrap) log.Logger {
	return f.KLogger()
}

func ExportZLogger(f *Bootstrap) *zap.Logger {
	return f.ZLogger()
}

func ExportSLogger(f *Bootstrap) *slog.Logger {
	return f.SLogger()
}

func CreateApp(c *Bootstrap, servers []transport.Server) *kratos.App {

	return kratos.New(
		kratos.ID(c.Meta.ServiceID),
		kratos.Name(c.Meta.ServiceName),
		kratos.Version(c.Meta.Version),
		kratos.Metadata(map[string]string{}),
		kratos.Logger(c.KLogger()),
		kratos.Server(
			servers...,
		),
	)
}

var ProviderSet = wire.NewSet(ExportSLogger, ExportZLogger, ExportLogger, CreateApp)
