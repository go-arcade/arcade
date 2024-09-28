package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/go-arcade/arcade/internal/app/engine/conf"
	"github.com/go-arcade/arcade/internal/app/engine/server"
	"github.com/go-arcade/arcade/internal/router"
	"github.com/go-arcade/arcade/pkg/ctx"
	"github.com/go-arcade/arcade/pkg/database"
	"github.com/go-arcade/arcade/pkg/log"
	"github.com/go-arcade/arcade/pkg/runner"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/9/4 19:51
 * @file: engine.go
 * @description: engine program
 */

var (
	configFile string
)

func init() {
	flag.StringVar(&configFile, "conf", "conf.d/config.toml", "conf file path, e.g. -conf ./conf.d")
}

func main() {
	flag.Parse()
	printRunner()

	var appConf conf.AppConfig
	appConf = conf.NewConf(configFile)

	logger := log.NewLog(&appConf.Log)

	//_, err := cache.NewRedis(appConf.Redis)
	//if err != nil {
	//	panic(err)
	//}

	// db
	db := database.NewDatabase(appConf.Database)

	Ctx := ctx.NewContext(context.Background(), db, logger)

	// repo.NewAgentRepo(db)

	route := router.NewRouter(&appConf.Http, Ctx)

	// httpx srv
	http := server.NewHttp(appConf.Http)
	httpClean := http.Server(route.Router())

	httpClean()
}

func printRunner() {
	fmt.Println("runner.pwd:", runner.Pwd)
	fmt.Println("runner.hostname:", runner.Hostname)
}
