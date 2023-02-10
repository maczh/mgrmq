package rabbitmq

import (
	"fmt"
	"github.com/levigross/grequests"
	"github.com/maczh/mgin/cache"
	"github.com/maczh/mgin/client"
	"github.com/maczh/mgin/logs"
	"github.com/maczh/mgin/utils"
	"github.com/maczh/mgrabbit"
	"github.com/maczh/mgrmq/model"
	"github.com/maczh/mgrmq/mongo"
	"gopkg.in/mgo.v2/bson"
	"net/url"
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
	mongo.NewFailLogMgo().Save(handler.Job.FailLog, failLog)
}

func (handler *QueueHandler) LogCall(callLog model.CallLog) {
	if handler.Job.ReqLog == "" {
		return
	}
	mongo.NewCallLogMgo().Save(handler.Job.ReqLog, callLog)
}

func (handler *QueueHandler) Listening(msg []byte) {
	m := string(msg)
	logs.Debug("队列{}收到消息:{}", handler.Job.Queue, m)
	if m == "" || m[:1] != "{" {
		logs.Error("消息内容不是JSON格式，无法解析")
		return
	}
	param := client.Options{}
	utils.FromJSON(m, &param)
	key := utils.MD5Encode(m)
	v, found := cache.OnGetCache("counters").Value(key)
	if found && v.(int) >= handler.Job.Retry {
		logs.Error("队列{}的消息{}已经处理{}次，停止处理", handler.Job.Queue, m, v.(int))
		cache.OnGetCache("counters").Delete(key)
		//记录失败日志
		handler.LogFailed(m)
		return
	} else if !found {
		v = 1
	} else {
		v = v.(int) + 1
	}
	cache.OnGetCache("counters").Add(key, v, time.Duration(handler.Job.Interval*2)*time.Second)
	begin := time.Now()
	callLog := model.CallLog{
		Id:       bson.NewObjectId(),
		Time:     utils.GetNowDateTime(),
		Queue:    handler.Job.Queue,
		Msg:      m,
		CallType: handler.Job.CallType,
		Params:   param,
		Service:  handler.Job.Service,
		Uri:      handler.Job.Uri,
		Url:      handler.Job.Url,
	}
	if handler.Job.CallType == "service" {
		if handler.Job.Service == "" {
			logs.Error("任务{}的微服务名配置不可为空", handler.Job.Name)
			return
		}
		if handler.Job.Uri == "" {
			logs.Error("任务{}的微服务接口地址uri配置不可为空", handler.Job.Name)
		}
		resp := client.CallT[any](handler.Job.Service, handler.Job.Uri, &param)
		end := time.Now()
		callLog.ResponseTime = utils.ToDateTimeString(end)
		callLog.Ttl = end.UnixMilli() - begin.UnixMilli()
		if resp.Status != 1 {
			logs.Error("微服务{}{}调用错误:{}", handler.Job.Service, handler.Job.Uri, resp.Msg)
			callLog.Response = resp.Msg
			handler.LogCall(callLog)
			if handler.Job.RetryOn == "failed" && v.(int) <= handler.Job.Retry {
				mgrabbit.Rabbit.RabbitSendMessage(handler.Job.QueueDx, m)
			}
			return
		}
		callLog.Response = resp
		handler.LogCall(callLog)
	} else if handler.Job.CallType == "http" {
		if handler.Job.Url == "" {
			logs.Error("任务{}的接口地址url不可为空", handler.Job.Name)
			return
		}
		if handler.Job.Method == "" {
			handler.Job.Method = param.Method
		}
		if handler.Job.ContentType == "" {
			switch param.Protocol {
			case client.CONTENT_TYPE_FORM, "":
				handler.Job.ContentType = "application/x-www-form-urlencoded"
			case client.CONTENT_TYPE_JSON:
				handler.Job.ContentType = "application/json"
			}
		}
		if handler.Job.Success == "" {
			logs.Error("任务{}的接口返回成功标志配置不可为空", handler.Job.Name)
		}
		uri := handler.Job.Url
		if handler.Job.ContentType == "application/x-www-form-urlencoded" && param.Path != nil && len(param.Path) > 0 {
			for k, v := range param.Path {
				uri = strings.ReplaceAll(uri, fmt.Sprintf("{%s}", k), url.PathEscape(v))
			}
		}
		var resp *grequests.Response
		var err error
		headers := utils.AnyToMap(param.Header)
		headers["Content-Type"] = handler.Job.ContentType
		switch strings.ToUpper(handler.Job.Method) {
		case "GET":
			resp, err = grequests.Get(uri, &grequests.RequestOptions{
				Params:  utils.AnyToMap(param.Query),
				Headers: headers,
			})
		case "POST":
			if strings.Contains(handler.Job.ContentType, "application/x-www-form-urlencoded") {
				resp, err = grequests.Post(uri, &grequests.RequestOptions{
					Headers: headers,
					Data:    utils.AnyToMap(param.Data),
				})
			} else { //json格式
				resp, err = grequests.Post(uri, &grequests.RequestOptions{
					Headers: headers,
					JSON:    param.Json,
				})
			}
		default:
			logs.Error("http模式仅支持GET与POST")
		}
		end := time.Now()
		callLog.ResponseTime = utils.ToDateTimeString(end)
		callLog.Ttl = end.UnixMilli() - begin.UnixMilli()
		if err != nil {
			logs.Error("接口{}调用错误:{}", uri, err.Error())
			callLog.Response = err.Error()
			handler.LogCall(callLog)
			if v.(int) <= handler.Job.Retry {
				mgrabbit.Rabbit.RabbitSendMessage(handler.Job.QueueDx, m)
			}
			return
		}
		if resp.StatusCode != 200 {
			logs.Error("接口{}调用错误:{}", handler.Job.Url, resp.StatusCode)
			callLog.Response = resp.String()
			handler.LogCall(callLog)
			if v.(int) <= handler.Job.Retry {
				mgrabbit.Rabbit.RabbitSendMessage(handler.Job.QueueDx, m)
			}
			return
		}
		logs.Debug("外部接口{}调用返回:{}", handler.Job.Url, resp.String())
		if resp.String()[:1] == "{" {
			result := make(map[string]interface{})
			utils.FromJSON(resp.String(), &result)
			callLog.Response = result
			handler.LogCall(callLog)
		} else {
			callLog.Response = resp.String()
			handler.LogCall(callLog)
		}
		if handler.Job.RetryOn == "failed" {
			if !strings.Contains(resp.String(), handler.Job.Success) {
				logs.Error("接口{}调用错误:{}", handler.Job.Url, resp.String())
				if v.(int) <= handler.Job.Retry {
					mgrabbit.Rabbit.RabbitSendMessage(handler.Job.QueueDx, m)
				}
			}
		}

	} else {
		logs.Error("callType配置只能是service或http")
	}
}
