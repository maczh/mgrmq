package controller

import (
	"encoding/json"
	"github.com/maczh/gintool/mgresult"
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
// @Param	msg formData string true "消息内容，必须是JSON格式"
// @Success 200 {string} string	"ok"
// @Router	/msg/send [post]
func SendMessage(params map[string]string) mgresult.Result {
	queue, msg := params["queue"], params["msg"]
	if queue == "" {
		return mgresult.Error(-1, "消息队列名称不可为空")
	}
	if msg == "" {
		return mgresult.Error(-1, "消息内容不可为空")
	}
	var m map[string]string
	err := json.Unmarshal([]byte(msg), &m)
	if err != nil {
		return mgresult.Error(-1, "消息内容不是JSON格式")
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
func ReSendFailedMessage(params map[string]string) mgresult.Result {
	return msgService.ReSend(params["queue"], params["start"], params["end"])
}
