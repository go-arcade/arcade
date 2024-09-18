package router

import (
	"github.com/gin-gonic/gin"
	"github.com/go-arcade/arcade/pkg/httpx/ws"
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

func EngineRouter(r *gin.RouterGroup) {

	//staticsFS, err := fs.New()
	//if err != nil {
	//	log.Errorf("cannot create statik fs: %v", err)
	//}
	//r.StaticFS("/static", staticsFS)

	r.POST("/ws", ws.Handle)

	route := r.Group("/agent")
	{
		route.POST("/add", addAgent)
		//r.POST("delete", deleteAgent)
		//r.POST("update", updateAgent)
		route.GET("/list", listAgent)
	}

}
