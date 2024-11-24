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
	httpx "github.com/go-arcade/arcade/pkg/http"
	"github.com/go-arcade/arcade/pkg/log"
	"github.com/go-arcade/arcade/pkg/runner"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/9/4 19:51
 * @file: main.go
 * @description: arcade program
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
	db, err := database.NewDatabase(appConf.Database)
	if err != nil {
		panic(err)
	}
	mongo, err := database.NewMongoDB(appConf.Database.MongoDB, context.Background())
	if err != nil {
		panic(err)
	}
	Ctx := ctx.NewContext(context.Background(), mongo, redis, db, logger)

	route := router.NewRouter(&appConf.Http, Ctx)
	// http srv
	cleanup := httpx.NewHttp(appConf.Http, route.Router())
	cleanup()
}

func printRunner() {
	fmt.Println("runner.pwd:", runner.Pwd)
	fmt.Println("runner.hostname:", runner.Hostname)
}
