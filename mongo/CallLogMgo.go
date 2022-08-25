package mongo

import (
	"github.com/maczh/logs"
	"github.com/maczh/mgconfig"
	"github.com/maczh/mgrmq/model"
)

type CallLogMgo struct {}

func NewCallLogMgo() *CallLogMgo  {
	return &CallLogMgo{}
}

func (f *CallLogMgo) Save(collectionName string, log model.CallLog) error {
	mongo,err := mgconfig.GetMongoConnection()
	if err != nil {
		logs.Error("MongoDB connection error:{}",err.Error())
		return err
	}
	defer mgconfig.ReturnMongoConnection(mongo)
	err = mongo.C(collectionName).Insert(log)
	return err
}
