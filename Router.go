package main

import (
	"github.com/ekyoung/gin-nice-recovery"
	"github.com/gin-gonic/gin"
	"github.com/maczh/mgin/middleware/cors"
	"github.com/maczh/mgin/middleware/postlog"
	"github.com/maczh/mgin/middleware/trace"
	"github.com/maczh/mgin/models"
	"github.com/maczh/mgrmq/controller"

	_ "github.com/maczh/mgrmq/docs"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"net/http"
)

/*
*
统一路由映射入口
*/
func setupRouter() *gin.Engine {
	// Disable Console Color
	// gin.DisableConsoleColor()
	engine := gin.Default()

	//添加跟踪日志
	engine.Use(trace.TraceId())

	//设置接口日志
	engine.Use(postlog.RequestLogger())
	//添加跨域处理
	engine.Use(cors.Cors())

	//添加swagger支持
	engine.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	//处理全局异常
	engine.Use(nice.Recovery(recoveryHandler))

	//设置404返回的内容
	engine.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusOK, models.Error(-1, "404 Not Found"))
	})

	var result models.Result[any]
	//添加所需的路由映射
	//消息处理
	engine.POST("/msg/send", func(c *gin.Context) {
		result = controller.SendMessage(c)
		c.JSON(http.StatusOK, result)
	})

	engine.POST("/msg/resend", func(c *gin.Context) {
		result = controller.ReSendFailedMessage(c)
		c.JSON(http.StatusOK, result)
	})

	return engine
}

func recoveryHandler(c *gin.Context, err interface{}) {
	c.JSON(http.StatusOK, models.Error(-1, "System Error"))
}
