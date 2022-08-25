package rabbitmq

import (
	"github.com/levigross/grequests"
	"github.com/maczh/gintool/mgresult"
	"github.com/maczh/logs"
	"github.com/maczh/mgcache"
	"github.com/maczh/mgcall"
	"github.com/maczh/mgconfig"
	"github.com/maczh/mgrmq/model"
	"github.com/maczh/mgrmq/mongo"
	"github.com/maczh/utils"
	"gopkg.in/mgo.v2/bson"
	"strings"
	"time"
)

type QueueHandler struct {
	Job model.MQjob
}

func NewQueueHandler(job model.MQjob) *QueueHandler {
	handler := &QueueHandler{
		Job: job,
	}
	return handler
}

func (handler *QueueHandler) LogFailed(msg string) {
	failLog := model.FailLog{
		Id:    bson.NewObjectId(),
		Time:  utils.GetNowDateTime(),
		Queue: handler.Job.Queue,
		Msg:   msg,
		Retry: handler.Job.Retry,
	}
	if handler.Job.FailLog == "" {
		return
	}
	mongo.NewFailLogMgo().Save(handler.Job.FailLog,failLog)
}

func (handler *QueueHandler) LogCall(callLog model.CallLog){
	if handler.Job.ReqLog == "" {
		return
	}
	mongo.NewCallLogMgo().Save(handler.Job.ReqLog,callLog)
}

func (handler *QueueHandler)Listening(msg []byte) {
	m := string(msg)
	logs.Debug("队列{}收到消息:{}",handler.Job.Queue,m)
	if m == "" || m[:1] != "{" {
		logs.Error("消息内容不是JSON格式，无法解析")
		return
	}
	param := make(map[string]string)
	utils.FromJSON(m,&param)
	key := utils.MD5Encode(m)
	v,found := mgcache.OnGetCache("counters").Value(key)
	if found && v.(int) >= handler.Job.Retry {
		logs.Error("队列{}的消息{}已经处理{}次，停止处理",handler.Job.Queue,m,v.(int))
		mgcache.OnGetCache("counters").Delete(key)
		//记录失败日志
		handler.LogFailed(m)
		return
	}else if !found {
		v = 1
	}else {
		v = v.(int) +1
	}
	mgcache.OnGetCache("counters").Add(key,v, time.Duration(handler.Job.Interval*2)*time.Second)
	begin := time.Now()
	callLog := model.CallLog{
		Id:       bson.NewObjectId(),
		Time:     utils.GetNowDateTime(),
		Queue:    handler.Job.Queue,
		Msg:      m,
		CallType: handler.Job.CallType,
		Params:   param,
		Service: handler.Job.Service,
		Uri: handler.Job.Uri,
		Url: handler.Job.Url,
	}
	if handler.Job.CallType == "service" {
		if handler.Job.Service == "" {
			logs.Error("任务{}的微服务名配置不可为空",handler.Job.Name)
			return
		}
		if handler.Job.Uri == "" {
			logs.Error("任务{}的微服务接口地址uri配置不可为空",handler.Job.Name)
		}
		resp,err := mgcall.Call(handler.Job.Service,handler.Job.Uri,param)
		end := time.Now()
		callLog.ResponseTime = utils.ToDateTimeString(end)
		callLog.Ttl = end.UnixMilli() - begin.UnixMilli()
		if err != nil {
			logs.Error("微服务{}{}调用错误:{}",handler.Job.Service,handler.Job.Uri,err.Error())
			callLog.Response = err.Error()
			handler.LogCall(callLog)
			if v.(int) <= handler.Job.Retry {
				mgconfig.RabbitSendMessage(handler.Job.QueueDx, m)
			}
			return
		}
		var result mgresult.Result
		utils.FromJSON(resp,&result)
		callLog.Response = result
		handler.LogCall(callLog)
		if handler.Job.RetryOn == "failed" {
			if result.Status != 1 {
				logs.Error("微服务{}{}调用错误:{}",handler.Job.Service,handler.Job.Uri,result.Msg)
				if v.(int) <= handler.Job.Retry {
					mgconfig.RabbitSendMessage(handler.Job.QueueDx, m)
				}
			}
		}
	}else if handler.Job.CallType == "http" {
		if handler.Job.Url == "" {
			logs.Error("任务{}的接口地址url不可为空",handler.Job.Name)
			return
		}
		if handler.Job.Method == "" {
			logs.Error("任务{}的http方法配置不可为空",handler.Job.Name)
		}
		if handler.Job.ContentType == "" {
			handler.Job.ContentType = "application/x-www-form-urlencoded"
		}
		if handler.Job.Success == "" {
			logs.Error("任务{}的接口返回成功标志配置不可为空",handler.Job.Name)
		}
		var resp *grequests.Response
		var err error
		switch strings.ToUpper(handler.Job.Method) {
		case "GET":
			resp,err = grequests.Get(handler.Job.Url,&grequests.RequestOptions{Params: param})
		case "POST":
			if strings.Contains(handler.Job.ContentType , "application/x-www-form-urlencoded") {
				resp,err = grequests.Post(handler.Job.Url,&grequests.RequestOptions{Data: param})
			}else {	//json格式
				resp,err = grequests.Post(handler.Job.Url,&grequests.RequestOptions{JSON: param})
			}
		default:
			logs.Error("http模式仅支持GET与POST")
		}
		end := time.Now()
		callLog.ResponseTime = utils.ToDateTimeString(end)
		callLog.Ttl = end.UnixMilli() - begin.UnixMilli()
		if err != nil {
			logs.Error("接口{}调用错误:{}",handler.Job.Url,err.Error())
			callLog.Response = err.Error()
			handler.LogCall(callLog)
			if v.(int) <= handler.Job.Retry {
				mgconfig.RabbitSendMessage(handler.Job.QueueDx, m)
			}
			return
		}
		if resp.StatusCode != 200 {
			logs.Error("接口{}调用错误:{}",handler.Job.Url,resp.StatusCode)
			callLog.Response = resp.String()
			handler.LogCall(callLog)
			if v.(int) <= handler.Job.Retry {
				mgconfig.RabbitSendMessage(handler.Job.QueueDx, m)
			}
			return
		}
		logs.Debug("外部接口{}调用返回:{}",handler.Job.Url,resp.String())
		if resp.String()[:1] == "{" {
			result := make(map[string]interface{})
			utils.FromJSON(resp.String(), &result)
			callLog.Response = result
			handler.LogCall(callLog)
		}else {
			callLog.Response = resp.String()
			handler.LogCall(callLog)
		}
		if handler.Job.RetryOn == "failed" {
			if !strings.Contains(resp.String(),handler.Job.Success) {
				logs.Error("接口{}调用错误:{}",handler.Job.Url,resp.String())
				if v.(int) <= handler.Job.Retry {
					mgconfig.RabbitSendMessage(handler.Job.QueueDx, m)
				}
			}
		}

	}else {
		logs.Error("callType配置只能是service或http")
	}
}