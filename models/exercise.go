// @Title 用户习题数据模型
// @Description 用户习题数据模型和操作方法
package models

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"maiyajia.com/services/mongo"
)

type ExerModels struct {
	MgoSession *mongo.MgoClient
}
type Exercise struct {
	ID          bson.ObjectId `bson:"_id" json:"id"`                  //ID
	CourseID    bson.ObjectId `bson:"courseID" json:"courseID"`       //课程ID
	LessonID    bson.ObjectId `bson:"lessonID" json:"lessonID"`       //课时ID
	UserID      bson.ObjectId `bson:"userID" json:"userID"`           //用户ID
	CouerseName string        `bson:"couerseName" json:"couerseName"` //课程名字
	LessonName  string        `bson:"lessonName" json:"lessonName"`   //课时名字
	Answer      string        `bson:"answer" json:"answer"`           //习题答案
}

func (exerMod *ExerModels) InsertExercise(exer Exercise) error {
	logs.Info(exer)
	f := func(col *mgo.Collection) error {
		return col.Insert(exer)
	}
	return exerMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "exercise", f)
}
func (exerMod *ExerModels) GetExercise(courseID bson.ObjectId) {

}
