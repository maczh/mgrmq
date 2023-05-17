package service

import (
	"github.com/maczh/mgin/config"
	"github.com/maczh/mgin/logs"
	"github.com/maczh/mgin/middleware/trace"
	"github.com/maczh/mgin/models"
	"github.com/maczh/mgrabbit"
	"github.com/maczh/mgrmq/mongo"
)

type MessageService struct {
	multi bool
	tag   string
}

func NewMessageService() *MessageService {
	return &MessageService{
		multi: config.Config.GetConfigBool("go.db.multi"),
		tag:   config.Config.GetConfigString("go.db.dbName"),
	}
}

func (s *MessageService) Send(queue, msg string) models.Result[any] {
	if _, ok := queues[queue]; !ok {
		return models.Error(-1, "消息队列名称不在配置中")
	}
	dbName := ""
	if s.multi {
		dbName = trace.GetHeader(s.tag)
	}
	conn, err := mgrabbit.Rabbit.GetConnection(dbName)
	if err != nil {
		logs.Error("获取RabbitMQ连接失败:{}", err.Error())
		return models.Error(-1, "消息队列连接失败:"+err.Error())
	}
	conn.RabbitSendMessage(queue, msg)
	return models.Success[any](nil)
}

func (s *MessageService) ReSend(queue, start, end string) models.Result[any] {
	dbName := ""
	if s.multi {
		dbName = trace.GetHeader(s.tag)
	}
	if queue == "" {
		for q, job := range jobConfigs {
			fails, err := mongo.NewFailLogMgo().List(dbName, job.FailLog, start, end)
			if err != nil {
				logs.Error("获取队列{}在{}到{}时间段内的失败日志失败:{}", q, start, end, err.Error())
				continue
			}
			for _, log := range fails {
				s.Send(log.Queue, log.Msg)
			}
		}
	} else {
		fails, err := mongo.NewFailLogMgo().List(dbName, jobConfigs[queue].FailLog, start, end)
		if err != nil {
			logs.Error("获取队列{}在{}到{}时间段内的失败日志失败:{}", queue, start, end, err.Error())
			return models.Error(-1, err.Error())
		}
		for _, log := range fails {
			s.Send(log.Queue, log.Msg)
		}
	}
	return models.Success[any](nil)
}
