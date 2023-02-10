package model

type SendReq struct {
	Queue string `json:"queue" form:"queue"`
	Msg   string `json:"msg" form:"msg"`
}

type ReSendReq struct {
	Queue string `json:"queue" form:"queue"`
	Start string `json:"start" form:"start"`
	End   string `json:"end" form:"end"`
}
