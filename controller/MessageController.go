package controller

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/maczh/mgin/client"
	"github.com/maczh/mgin/logs"
	"github.com/maczh/mgin/models"
	"github.com/maczh/mgrmq/model"
	"github.com/maczh/mgrmq/service"
)

var msgService = service.NewMessageService()

// SendMessage	godoc
// @Summary		发送一条消息到消息队列
// @Description	发送一条消息到消息队列
// @Tags	发布
// @Accept	x-www-form-urlencoded
// @Produce json
// @Param	queue formData string true "消息队列名"
// @Param	msg formData string true "消息内容，必须是mgin中的client.Options格式"
// @Success 200 {string} string	"ok"
// @Router	/msg/send [post]
func SendMessage(c *gin.Context) models.Result[any] {
	var req model.SendReq
	var err error
	switch c.ContentType() {
	case gin.MIMEJSON:
		err = c.ShouldBindJSON(&req)
	default:
		err = c.ShouldBind(&req)
	}
	if err != nil {
		logs.Error("绑定入参错误:{}", err.Error())
		return models.Error(-1, "入参模式错误")
	}
	queue, msg := req.Queue, req.Msg
	if queue == "" {
		return models.Error(-1, "消息队列名称不可为空")
	}
	if msg == "" {
		return models.Error(-1, "消息内容不可为空")
	}
	var msgBody client.Options
	err = json.Unmarshal([]byte(msg), &msgBody)
	if err != nil || msgBody.Method == "" {
		return models.Error(-1, "消息内容不正确")
	}

	return msgService.Send(queue, msg)
}

// ReSendFailedMessage	godoc
// @Summary		重新发送指定时间段内的失败消息
// @Description	重新发送指定时间段内的失败消息
// @Tags	发布
// @Accept	x-www-form-urlencoded
// @Produce json
// @Param	queue formData string false "消息队列名，不传为配置中的所有队列"
// @Param	start formData string false "开始时间，格式 yyyy-MM-dd HH:mm:ss,不传默认为当天0点"
// @Param	end formData string false "结束时间，格式 yyyy-MM-dd HH:mm:ss,不传默认为当前时间"
// @Success 200 {string} string	"ok"
// @Router	/msg/resend [post]
func ReSendFailedMessage(c *gin.Context) models.Result[any] {
	var req model.ReSendReq
	var err error
	switch c.ContentType() {
	case gin.MIMEJSON:
		err = c.ShouldBindJSON(&req)
	default:
		err = c.ShouldBind(&req)
	}
	if err != nil {
		logs.Error("绑定入参错误:{}", err.Error())
		return models.Error(-1, "入参模式错误")
	}
	return msgService.ReSend(req.Queue, req.Start, req.End)
}
