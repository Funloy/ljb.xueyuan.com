package models

import (
	"time"

	"github.com/astaxie/beego"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"maiyajia.com/services/mongo"
)

type LoggerModels struct {
	MgoSession *mongo.MgoClient
}

// PassportLog 登录日志
type PassportLog struct {
	UserID      bson.ObjectId `bson:"userID" json:"userID"`
	LoginIP     string        `bson:"loginIP" json:"loginIP"`
	LoginProxy  string        `bson:"loginProxy" json:"loginProxy"`
	LoginClient string        `bson:"loginClient" json:"loginClinet"`
	Action      string        `bson:"action" json:"action"`
	LogTime     time.Time     `bson:"logTime" json:"logTime"` //记录时间
}

// Logger 用户日志
func (logMod *LoggerModels) Logger(log interface{}) error {
	f := func(col *mgo.Collection) error {
		return col.Insert(log)
	}
	err := logMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "log", f)
	return err
}
