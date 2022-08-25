package mongo

import (
	"github.com/maczh/logs"
	"github.com/maczh/mgconfig"
	"github.com/maczh/mgrmq/model"
	"github.com/maczh/utils"
	"gopkg.in/mgo.v2/bson"
	"time"
)

type FailLogMgo struct {}

func NewFailLogMgo() *FailLogMgo  {
	return &FailLogMgo{}
}

func (f *FailLogMgo) Save(collectionName string, log model.FailLog) error {
	mongo,err := mgconfig.GetMongoConnection()
	if err != nil {
		logs.Error("MongoDB connection error:{}",err.Error())
		return err
	}
	defer mgconfig.ReturnMongoConnection(mongo)
	err = mongo.C(collectionName).Insert(log)
	return err
}

func (f *FailLogMgo) List(collection, start, end string) ([]model.FailLog, error) {
	mongo,err := mgconfig.GetMongoConnection()
	if err != nil {
		logs.Error("MongoDB connection error:{}",err.Error())
		return nil,err
	}
	defer mgconfig.ReturnMongoConnection(mongo)
	var logs []model.FailLog
	if start == "" {
		start = utils.ToDateTimeString(time.Now())
	}
	timqQuery := bson.M{
		"$gt": start,
	}
	if end != "" {
		timqQuery["$lt"] = end
	}
	query := bson.M{
		"time" : timqQuery,
	}
	err = mongo.C(collection).Find(query).All(&logs)
	return logs,err
}