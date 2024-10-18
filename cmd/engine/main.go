package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/go-arcade/arcade/internal/engine/conf"
	"github.com/go-arcade/arcade/internal/engine/router"
	"github.com/go-arcade/arcade/pkg/cache"
	"github.com/go-arcade/arcade/pkg/ctx"
	"github.com/go-arcade/arcade/pkg/database"
	"github.com/go-arcade/arcade/pkg/http"
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

	redis, err := cache.NewRedis(appConf.Redis)
	if err != nil {
		panic(err)
	}

	// db
	db := database.NewDatabase(appConf.Database)
	mongodb := database.NewMongoDB(appConf.Database.MongoDB)
	mongoIns, err := mongodb.Connect(context.Background())
	if err != nil {
		panic(err)
	}
	Ctx := ctx.NewContext(context.Background(), mongoIns, redis, db, logger)

	route := router.NewRouter(&appConf.Http, Ctx)

	// http srv
	http := http.NewHttp(appConf.Http)
	httpClean := http.Server(route.Router())

	httpClean()
}

func printRunner() {
	fmt.Println("runner.pwd:", runner.Pwd)
	fmt.Println("runner.hostname:", runner.Hostname)
}
