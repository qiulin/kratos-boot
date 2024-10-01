//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"github.com/google/wire"
	"github.com/qiulin/kratos-boot/boot"
	"github.com/qiulin/kratos-boot/console"
)

// wireApp init kratos application.
func wireApp(*boot.Bootstrap) (*console.CommandLine, func(), error) {
	panic(wire.Build(ProviderSet))
}
