package mongo

import (
	"github.com/maczh/mgin/db"
	"github.com/maczh/mgin/logs"
	"github.com/maczh/mgrmq/model"
)

type CallLogMgo struct{}

func NewCallLogMgo() *CallLogMgo {
	return &CallLogMgo{}
}

func (f *CallLogMgo) Save(collectionName string, log model.CallLog) error {
	mongo, err := db.Mongo.GetConnection()
	if err != nil {
		logs.Error("MongoDB connection error:{}", err.Error())
		return err
	}
	defer db.Mongo.ReturnConnection(mongo)
	err = mongo.C(collectionName).Insert(log)
	return err
}
