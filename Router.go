package main

import (
	"github.com/ekyoung/gin-nice-recovery"
	"github.com/gin-gonic/gin"
	"github.com/maczh/gintool"
	"github.com/maczh/gintool/mgresult"
	"github.com/maczh/mgrmq/controller"
	"github.com/maczh/mgtrace"
	"github.com/maczh/utils"

	_ "github.com/maczh/mgrmq/docs"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"
	"net/http"
)

/**
统一路由映射入口
*/
func setupRouter() *gin.Engine {
	// Disable Console Color
	// gin.DisableConsoleColor()
	engine := gin.Default()

	//添加跟踪日志
	engine.Use(mgtrace.TraceId())

	//设置接口日志
	engine.Use(gintool.SetRequestLogger())
	//添加跨域处理
	engine.Use(gintool.Cors())
	//添加国际化支持
	//engine.Use(mgerr.RequestLanguage())

	//添加swagger支持
	engine.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	//处理全局异常
	engine.Use(nice.Recovery(recoveryHandler))

	//设置404返回的内容
	engine.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusOK, mgresult.Error(-1, "404 Not Found"))
	})

	var result mgresult.Result
	//添加所需的路由映射
	//消息处理
	engine.POST("/msg/send", func(c *gin.Context) {
		result = controller.SendMessage(utils.GinParamMap(c))
		c.JSON(http.StatusOK, result)
	})

	engine.POST("/msg/resend", func(c *gin.Context) {
		result = controller.ReSendFailedMessage(utils.GinParamMap(c))
		c.JSON(http.StatusOK, result)
	})

	return engine
}

func recoveryHandler(c *gin.Context, err interface{}) {
	c.JSON(http.StatusOK, mgresult.Error(-1, "System Error"))
}
