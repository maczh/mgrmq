package model

import "gopkg.in/mgo.v2/bson"

type CallLog struct {
	Id           bson.ObjectId     `json:"id" bson:"_id"`
	Time         string            `json:"time" bson:"time"`
	ResponseTime string            `json:"responseTime" bson:"responseTime"`
	Ttl          int64             `json:"ttl" bson:"ttl"`
	Queue        string            `json:"queue" bson:"queue"`
	Msg          string            `json:"msg" bson:"msg"`
	CallType     string            `json:"callType" bson:"callType"`
	Service      string            `json:"service,omitempty" bson:"service,omitempty"`
	Uri          string            `json:"uri,omitempty" bson:"uri,omitempty"`
	Url          string            `json:"url,omitempty" bson:"url,omitempty"`
	Params       map[string]string `json:"params" bson:"params"`
	Response     interface{}       `json:"response" bson:"response"`
}

type FailLog struct {
	Id    bson.ObjectId `json:"id" bson:"_id"`
	Time  string        `json:"time" bson:"time"`
	Queue string        `json:"queue" bson:"queue"`
	Msg   string        `json:"msg" bson:"msg"`
	Retry int           `json:"retry" bson:"retry"`
}
