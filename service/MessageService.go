package service

import (
	"github.com/maczh/mgin/logs"
	"github.com/maczh/mgin/models"
	"github.com/maczh/mgrabbit"
	"github.com/maczh/mgrmq/mongo"
)

type MessageService struct{}

func NewMessageService() *MessageService {
	return &MessageService{}
}

func (s *MessageService) Send(queue, msg string) models.Result[any] {
	if _, ok := queues[queue]; !ok {
		return models.Error(-1, "消息队列名称不在配置中")
	}
	mgrabbit.Rabbit.RabbitSendMessage(queue, msg)
	return models.Success[any](nil)
}

func (s *MessageService) ReSend(queue, start, end string) models.Result[any] {
	if queue == "" {
		for q, job := range jobConfigs {
			fails, err := mongo.NewFailLogMgo().List(job.FailLog, start, end)
			if err != nil {
				logs.Error("获取队列{}在{}到{}时间段内的失败日志失败:{}", q, start, end, err.Error())
				continue
			}
			for _, log := range fails {
				s.Send(log.Queue, log.Msg)
			}
		}
	} else {
		fails, err := mongo.NewFailLogMgo().List(jobConfigs[queue].FailLog, start, end)
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
