package main

import (
	"flag"
	"fmt"
	"github.com/arcade/arcade/internal/app/basic/config"
	"github.com/arcade/arcade/internal/server/http"
	"github.com/arcade/arcade/pkg/cache"
	"github.com/arcade/arcade/pkg/conf"
	_ "github.com/arcade/arcade/pkg/conf"
	"os"
	"os/signal"
	"syscall"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/9/4 19:51
 * @file: core.go
 * @description: core program
 */

var (
	configFile  string
	releaseMode string
)

func init() {
	flag.StringVar(&configFile, "conf", "conf.d", "config file path, e.g. -conf ./conf.d")

	flag.StringVar(&releaseMode, "mode", "release", "http run mode")
}

func main() {
	flag.Parse()
	cfg := config.Config{}

	c, err := conf.LoadConfigFile(configFile, cfg)
	if err != nil {
		panic(err)
	}

	r := http.NewHTTPEngine(c.(config.Config).Http, releaseMode)

	httpClean := http.NewHTTP(c.(config.Config).Http, r)

	_, err = cache.NewRedis(c.(config.Config).Redis)
	if err != nil {
		panic(err)
	}

	code := 1
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

EXIT:
	for {
		sig := <-sc
		fmt.Println("received signal:", sig.String())
		switch sig {
		case syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
			code = 0
			break EXIT
		case syscall.SIGHUP:
			// todo: reload? or other?
		default:
			break EXIT
		}
	}

	httpClean()
	fmt.Println("server exit...")
	os.Exit(code)
}
