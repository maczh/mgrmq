package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/maczh/logs"
	"github.com/maczh/mgconfig"
	"github.com/maczh/mgrmq/service"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

const config_file = "mgrmq.yml"

//@title	RabbitMQ通用消息中转与处理模块
//@version 	1.0.0(mgrmq)
//@description	RabbitMQ通用消息中转与处理模块

func main() {
	//初始化配置，自动连接数据库和Nacos服务注册
	path, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	mgconfig.InitConfig(path + "/" + config_file)

	//GIN的模式，生产环境可以设置成release
	gin.SetMode("debug")

	//任务服务初始化
	service.NewJobService().Init()

	engine := setupRouter()

	server := &http.Server{
		Addr:    ":" + mgconfig.GetConfigString("go.application.port"),
		Handler: engine,
	}

	//common.PrintLogo()
	logs.Info("|-----------------------------------|")
	logs.Info("|             MGRMQ 1.0.0           |")
	logs.Info("|-----------------------------------|")
	logs.Info("|  Go Http Server Start Successful  |")
	logs.Info("|    Port:" + mgconfig.GetConfigString("go.application.port") + "     Pid:" + fmt.Sprintf("%d", os.Getpid()) + "        |")
	logs.Info("|-----------------------------------|")

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logs.Error("HTTP server listen: " + err.Error())
		}
	}()


	// 等待中断信号以优雅地关闭服务器（设置 5 秒的超时时间）
	signalChan := make(chan os.Signal)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGQUIT)
	sig := <-signalChan
	logs.Error("Get Signal:" + sig.String())
	logs.Error("Shutdown Server ...")
	mgconfig.SafeExit()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logs.Error("Server Shutdown:" + err.Error())
	}
	logs.Error("Server exiting")

}
