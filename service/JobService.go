package service

import (
	"github.com/levigross/grequests"
	"github.com/maczh/logs"
	"github.com/maczh/mgconfig"
	"github.com/maczh/mgrmq/model"
	"github.com/maczh/mgrmq/rabbitmq"
	"github.com/nacos-group/nacos-sdk-go/common/logger"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"path/filepath"
)

type JobService struct{}

func NewJobService() *JobService {
	return &JobService{}
}

var jobConfigs = make(map[string]model.MQjob)
var queues = make(map[string]string)

func loadConfig() (model.MQjobConfig, error) {
	prefix := mgconfig.GetConfigString("rmq.config.prefix")
	ymlData := []byte{}
	var err error
	if mgconfig.GetConfigString("rmq.config.source") == "file" {
		ymlFile := prefix + mgconfig.GetConfigString("go.config.mid") + mgconfig.GetConfigString("go.config.env") + mgconfig.GetConfigString("go.config.type")
		path, _ := filepath.Abs(filepath.Dir(os.Args[0]))
		ymlFile = path + "/" + ymlFile
		ymlData, err = ioutil.ReadFile(ymlFile)
		if err != nil {
			logs.Error("读取本地配置文件{}失败:{}", ymlFile, err.Error())
			return model.MQjobConfig{}, err
		}
	} else {
		configUrl := getConfigUrl(prefix)
		logs.Debug("正在获取mgrmq运行配置:{} ", configUrl)
		resp, err := grequests.Get(configUrl, nil)
		if err != nil {
			logger.Error("mgrmq配置下载失败! " + err.Error())
			return model.MQjobConfig{}, err
		}
		ymlData = resp.Bytes()
	}
	var jobConfig model.MQjobConfig
	err = yaml.Unmarshal(ymlData, &jobConfig)
	if err != nil {
		logs.Error("yaml配置解析错误:{}", err.Error())
		return jobConfig, err
	}
	logs.Debug("mgrmq配置解析结果:{}", jobConfig)
	for _, job := range jobConfig.Mgrmq.Jobs {
		jobConfigs[job.Queue] = job
		queues[job.Queue] = "1"
		if job.QueueDx != "" {
			queues[job.QueueDx] = "1"
		}
	}
	return jobConfig, err
}

func getConfigUrl(prefix string) string {
	serverType := mgconfig.GetConfigString("go.config.server_type")
	configUrl := mgconfig.GetConfigString("go.config.server")
	switch serverType {
	case "nacos":
		configUrl = configUrl + "nacos/v1/cs/configs?group=DEFAULT_GROUP&dataId=" + prefix + mgconfig.GetConfigString("go.config.mid") + mgconfig.GetConfigString("go.config.env") + mgconfig.GetConfigString("go.config.type")
	case "consul":
		configUrl = configUrl + "v1/kv/" + prefix + mgconfig.GetConfigString("go.config.mid") + mgconfig.GetConfigString("go.config.env") + mgconfig.GetConfigString("go.config.type") + "?dc=dc1&raw=true"
	case "springconfig":
		configUrl = configUrl + prefix + mgconfig.GetConfigString("go.config.mid") + mgconfig.GetConfigString("go.config.env") + mgconfig.GetConfigString("go.config.type")
	default:
		configUrl = configUrl + prefix + mgconfig.GetConfigString("go.config.mid") + mgconfig.GetConfigString("go.config.env") + mgconfig.GetConfigString("go.config.type")
	}
	return configUrl
}

func (js *JobService) Init() {
	jobConfig, err := loadConfig()
	if err != nil {
		logs.Error("加载配置失败:{}", err.Error())
		return
	}
	if len(jobConfig.Mgrmq.Jobs) == 0 {
		logs.Error("队列侦听任务配置错误或无任务配置")
		return
	}
	for _, job := range jobConfig.Mgrmq.Jobs {
		logs.Debug("正在初始化任务:{}", job.Name)
		if job.QueueDx == "" {
			logs.Error("死信队列名queueDx配置为空")
			continue
		}
		if job.Queue == "" {
			logs.Error("消息队列名queue配置为空")
			continue
		}
		if job.Interval < 1 {
			logs.Error("死信队列延时interval值不能小于1，单位是秒")
		}
		mgconfig.RabbitCreateDeadLetterQueue(job.QueueDx, job.Queue, job.Interval*1000)
		handler := rabbitmq.NewQueueHandler(job)
		mgconfig.RabbitMessageListener(job.Queue, handler.Listening)
		logs.Debug("正在侦听{}队列", job.Queue)
	}
	logs.Debug("所有队列侦听任务初始化均已完成")
}
