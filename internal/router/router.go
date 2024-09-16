package router

import (
	"github.com/gin-gonic/gin"
	_ "github.com/go-kratos/kratos/v2/log"
	_ "github.com/rakyll/statik/fs"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/9/8 15:48
 * @file: router.go
 * @description: router
 */

type Router struct {
}

func NewRouter(r *gin.Engine) {

	//staticsFS, err := fs.New()
	//if err != nil {
	//	log.Errorf("cannot create statik fs: %v", err)
	//}
	//r.StaticFS("/static", staticsFS)
}
