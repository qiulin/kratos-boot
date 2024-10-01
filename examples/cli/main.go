package main

import (
	"emperror.dev/emperror"
	"flag"
	"github.com/qiulin/kratos-boot/boot"
	"github.com/qiulin/kratos-boot/console"
	"os"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	// Name is the name of the compiled software.
	Name string
	// Version is the version of the compiled software.
	Version string
	// flagconf is the config flag.
	flagconf string

	id, _ = os.Hostname()
)

func init() {
	flag.StringVar(&flagconf, "conf", "examples/cli/configs", "config path, eg: -conf config.yaml")
}

func main() {
	bootstrap, cleanup, err := boot.NewBootstrap(&boot.Options{
		ConfigPath:      flagconf,
		ServiceId:       id,
		ServiceName:     Name,
		Version:         Version,
		ServiceMetadata: nil,
		EnvPrefix:       "CLI_",
	})
	emperror.Panic(err)
	defer cleanup()
	console.Execute(bootstrap, wireApp)
}
