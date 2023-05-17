package mongo

import (
	"github.com/maczh/mgin/db"
	"github.com/maczh/mgin/logs"
	"github.com/maczh/mgin/utils"
	"github.com/maczh/mgrmq/model"
	"gopkg.in/mgo.v2/bson"
	"time"
)

type FailLogMgo struct{}

func NewFailLogMgo() *FailLogMgo {
	return &FailLogMgo{}
}

func (f *FailLogMgo) Save(dbName, collectionName string, log model.FailLog) error {
	mongo, err := db.Mongo.GetConnection(dbName)
	if err != nil {
		logs.Error("MongoDB connection error:{}", err.Error())
		return err
	}
	defer db.Mongo.ReturnConnection(mongo)
	err = mongo.C(collectionName).Insert(log)
	return err
}

func (f *FailLogMgo) List(dbName, collection, start, end string) ([]model.FailLog, error) {
	mongo, err := db.Mongo.GetConnection(dbName)
	if err != nil {
		logs.Error("MongoDB connection error:{}", err.Error())
		return nil, err
	}
	defer db.Mongo.ReturnConnection(mongo)
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
		"time": timqQuery,
	}
	err = mongo.C(collection).Find(query).All(&logs)
	return logs, err
}
