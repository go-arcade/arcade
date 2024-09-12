//go:build wireinject
// +build wireinject

package main

import (
	"github.com/go-kratos/kratos/v2"
	"github.com/google/wire"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/9/8 20:44
 * @file: wire.go
 * @description:
 */

func initApp() (*kratos.App, func(), error) {
	panic(wire.Build())
}
