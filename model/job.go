package model

type MQjob struct {
	Name        string `json:"name" yaml:"name"`               //任务名称
	Queue       string `json:"queue" yaml:"queue"`             //队列名
	Retry       int    `json:"retry" yaml:"retry"`             //重试次数
	Interval    int    `json:"interval" yaml:"interval"`       //间隔时间
	QueueDx     string `json:"queueDx" yaml:"queueDx"`         //死信队列名
	FailLog     string `json:"failLog" yaml:"failLog"`         //处理失败记录表名，存于MongoDB
	ReqLog      string `json:"reqLog" yaml:"reqLog"`           //第三方接口请求日志
	CallType    string `json:"callType" yaml:"callType"`       //调用类型 service-微服务 http-直接http/https调用
	Service     string `json:"service" yaml:"service"`         //调用的微服务名
	Uri         string `json:"uri" yaml:"uri"`                 //微服务接口路径
	Url         string `json:"url" yaml:"url"`                 //http请求完整接口地址
	Method      string `json:"method" yaml:"method"`           //请求方法 GET/POST/PUT/DELETE
	ContentType string `json:"contentType" yaml:"contentType"` //http请求类型
	RetryOn     string `json:"retryOn" yaml:"retryOn"`         //重试条件 error-接口不通时 failed-接口返回失败时
	Success     string `json:"success" yaml:"success"`         //http请求返回成功的内容特征
}

type MQjobConfig struct {
	Mgrmq struct {
		Jobs      []MQjob `json:"jobs" yaml:"jobs"`
		ParamType string  `json:"paramType" yaml:"paramType"` //参数类型 map或options
	} `json:"mgrmq" yaml:"mgrmq"`
}

func (job *MQjob) Listen() {

}
