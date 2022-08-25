# RabbitMQ消息队列通用接收处理与重试服务

## 说明
- 本工程依赖mgconfig框架
- mgrmq通过配置自动侦听RabbitMQ的消息队列，收到消息后自动调用微服务接口或第三方接口，并且判断是否失败进行重试
- 使用本项目需在RabbitMQ中预先创建vhost，用户，授权及exchange，不需要预先创建队列，程序启动时将根据配置自动创建队列
- 若修改了配置中队列名称、死信队列名、死信队列ttl超时设置等，需在RabbitMQ中先删除相应队列后重启程序方可生效

## 本地特殊配置

```yaml
rmq:
  config:
    prefix: mgrmq   #文件名前缀
    source: nacos    #配置来源 file-本地文件 nacos-配置中心
```

## 队列任务配置范例

```yaml
mgrmq:
  jobs:
    - queue: q1                 #消息队列名称
      name: 队列1                #任务名称
      retry: 3                  #重试次数
      interval: 5               #死信队列延时，单位为秒
      queueDx: q1_d             #死信队列名称
      callType: service         #调用类型  service-微服务  http-http或https接口调用
      service: openapi-zone     #微服务的服务名
      uri: /api/zone/get        #微服务接口路径
      failLog: q1FailLog        #连续重试彻底失败后记录失败的表名称
      reqLog: q1CallLog         #接口调用日志表名
      retryOn: error            #重试条件 error-接口连接错误  failed-接口未返回成功标志
    - queue: q2
      name: 队列2
      retry: 2
      interval: 20
      queueDx: q2_d
      callType: http
      url: https://white.cdncloud.com:8442/api/v1/oss/get      #http接口完整地址
      method: POST                                              #接口请求方法 GET/POST
      contentType: application/x-www-form-urlencoded            #接口内容类型，支持application/x-www-form-urlencoded与application/json两种
      success: "\"status\":1,"                                  #接口执行成功标识
      failLog: q2FailLog
      reqLog: q2CallLog
      retryOn: failed
    - queue: q3
      name: 队列3
      retry: 3
      interval: 60
      queueDx: q3_d
      callType: service
      service: oss-server
      uri: /api/v1/oss/list
      failLog: q3FailLog
      reqLog: q3CallLog
      retryOn: error
```