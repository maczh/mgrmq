package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/maczh/mgin"
	"github.com/maczh/mgin/config"
	"github.com/maczh/mgin/logs"
	"github.com/maczh/mgrabbit"
	"github.com/maczh/mgrmq/controller"
	"github.com/maczh/mgrmq/service"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

const config_file = "mgrmq.yml"

// 初始化命令行参数
func parseArgs() string {
	var configFile string
	flag.StringVar(&configFile, "f", os.Args[0]+".yml", "yml配置文件名")
	flag.Parse()
	path, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	if !strings.Contains(configFile, "/") {
		configFile = path + "/" + configFile
	}
	return configFile
}

// @title	RabbitMQ通用消息中转与处理模块
// @version 	1.0.0(mgrmq)
// @description	RabbitMQ通用消息中转与处理模块
func main() {
	//初始化配置，自动连接数据库和Nacos服务注册
	configFile := parseArgs()
	mgin.Init(configFile)

	//GIN的模式，生产环境可以设置成release
	gin.SetMode("debug")

	mgin.MGin.Use("rabbitmq", mgrabbit.Rabbit.Init, mgrabbit.Rabbit.Close, nil)

	//任务服务初始化
	controller.Init()
	service.NewJobService().Init()

	engine := setupRouter()

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.Config.App.Port),
		Handler: engine,
	}

	//common.PrintLogo()
	logs.Info("|-----------------------------------|")
	logs.Info("|             MGRMQ 1.0.0           |")
	logs.Info("|-----------------------------------|")
	logs.Info("|  Go Http Server Start Successful  |")
	logs.Info("|    Port:" + config.Config.GetConfigString("go.application.port") + "     Pid:" + fmt.Sprintf("%d", os.Getpid()) + "        |")
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
	mgin.MGin.SafeExit()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logs.Error("Server Shutdown:" + err.Error())
	}
	logs.Error("Server exiting")

}
