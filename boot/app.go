package boot

import (
	"emperror.dev/emperror"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/env"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/google/wire"
	"github.com/qiulin/kratos-boot/discovery"
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
	opt     *Options
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

type Options struct {
	ConfigPath, ServiceId, ServiceName, Version string
	EnvPrefix                                   string
	ServiceMetadata                             map[string]string
}

func (opt *Options) EnsureDefaults() {

}

func defaultLogConfig() *sharedconf.Logging {
	return &sharedconf.Logging{
		Level: "debug",
	}
}

func NewBootstrap(opt *Options) (*Bootstrap, func(), error) {
	opt.EnsureDefaults()
	cs := config.New(
		config.WithSource(
			env.NewSource(opt.EnvPrefix),
			file.NewSource(opt.ConfigPath),
		))
	if err := cs.Load(); err != nil {
		return nil, nil, err
	}

	clog := &sharedconf.Logging{}
	if err := cs.Value("logging").Scan(clog); err != nil {
		if err == config.ErrNotFound {
			clog = defaultLogConfig()
		} else {
			return nil, nil, err
		}
	}

	zlogger, _ := logging.NewZapLogger(clog)
	zlogger = zlogger.With(zap.String("service_id", opt.ServiceId), zap.String("service_name", opt.ServiceName), zap.String("version", opt.Version))
	logger := logging.NewLogger(zlogger)
	slogger := logging.NewSLogger(clog.Level, zlogger)
	return &Bootstrap{
			klogger: logger,
			zlogger: zlogger,
			slogger: slogger,
			C:       cs,
			opt:     opt,
		}, func() {
			cs.Close()
		}, nil
}

func ExportLogger(b *Bootstrap) log.Logger {
	return b.KLogger()
}

func ExportZLogger(b *Bootstrap) *zap.Logger {
	return b.ZLogger()
}

func ExportSLogger(b *Bootstrap) *slog.Logger {
	return b.SLogger()
}

func CreateApp(b *Bootstrap, servers []transport.Server, f *discovery.Factory) *kratos.App {
	opts := []kratos.Option{
		kratos.ID(b.opt.ServiceId),
		kratos.Name(b.opt.ServiceName),
		kratos.Version(b.opt.Version),
		kratos.Metadata(b.opt.ServiceMetadata),
		kratos.Logger(b.KLogger()),
		kratos.Server(
			servers...,
		),
	}
	if r, exists := f.Registrar(); exists {
		opts = append(opts, kratos.Registrar(r))
	}

	return kratos.New(
		opts...,
	)
}

func ExportConfig(b *Bootstrap) config.Config {
	return b.C
}

type WireFunc func(bootstrap *Bootstrap) (*kratos.App, func(), error)

func RunOrPanic(b *Bootstrap, wireFunc WireFunc) {
	app, cleanup, err := wireFunc(b)
	if err != nil {
		panic(err)
	}
	defer cleanup()
	emperror.Panic(app.Run())
}

var ProviderSet = wire.NewSet(ExportSLogger, ExportZLogger, ExportLogger, CreateApp, ExportConfig, discovery.NewFactory)
