package main

import (
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/google/wire"
	"github.com/qiulin/kratos-boot/boot"
	"github.com/qiulin/kratos-boot/console"
	"github.com/qiulin/kratos-boot/examples/cli/conf"
)

//go:generate wire

func ExportConfig(b *boot.Bootstrap) *conf.Bootstrap {
	c := &conf.Bootstrap{}
	if err := b.ScanConfig(c); err != nil {
		panic(err)
	}
	return c
}

func NewConsoleOption() *console.Option {
	return &console.Option{
		Name: "hello cli",
	}
}

func NewServers(hs *http.Server) []transport.Server {
	return []transport.Server{hs}
}

func NewHTTPServer(bc *conf.Bootstrap) *http.Server {
	c := bc.Server
	var opts = []http.ServerOption{
		http.Middleware(
			recovery.Recovery(),
		),
	}
	if c.Http.Network != "" {
		opts = append(opts, http.Network(c.Http.Network))
	}
	if c.Http.Addr != "" {
		opts = append(opts, http.Address(c.Http.Addr))
	}
	if c.Http.Timeout != nil {
		opts = append(opts, http.Timeout(c.Http.Timeout.AsDuration()))
	}

	s := http.NewServer(opts...)
	return s
}

var ProviderSet = wire.NewSet(
	console.ProviderSet,
	NewHTTPServer,
	NewServers,
	NewConsoleOption,
	ExportConfig,
	NewHelloCommand,
	NewTimestampCommand,
	NewCommands,
)
