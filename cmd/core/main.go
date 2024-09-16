package main

import (
	"flag"
	"fmt"
	"github.com/arcade/arcade/internal/app/basic/config"
	"github.com/arcade/arcade/internal/router"
	"github.com/arcade/arcade/internal/server/http"
	"github.com/arcade/arcade/pkg/conf"
	"github.com/arcade/arcade/pkg/log"
	"github.com/arcade/arcade/pkg/orm"
	"github.com/arcade/arcade/pkg/runner"
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
	configFile string
)

func init() {
	flag.StringVar(&configFile, "conf", "conf.d", "config file path, e.g. -conf ./conf.d")
}

func main() {
	flag.Parse()
	printRunner()

	var appConf config.AppConfig
	if err := conf.LoadConfigFile(configFile, &appConf); err != nil {
		panic(err)
	}

	log.NewLog(&appConf.Log)

	//_, err := cache.NewRedis(appConf.Redis)
	//if err != nil {
	//	panic(err)
	//}

	// db
	orm.NewDatabase(appConf.Database)

	// http server
	r := http.NewHTTPEngine(appConf.Http)

	// router
	router.NewRouter(r)

	// http server clean
	httpClean := http.NewHTTP(appConf.Http, r)

	code := 1
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

EXIT:
	for {
		sig := <-sc
		fmt.Println("[Done] received signal:", sig.String())
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
	fmt.Println("[Done] server exit...")
	os.Exit(code)
}

func printRunner() {
	fmt.Println("runner.pwd:", runner.Pwd)
	fmt.Println("runner.hostname:", runner.Hostname)
}
