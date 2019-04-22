// @Title 用户学习进度数据模型
// @Description 用户学习进度的数据模型和操作方法

package models

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"maiyajia.com/services/mongo"
)

type ProgressModels struct {
	MgoSession *mongo.MgoClient
	CourseMod  CourseModels
	UserMod    UserModels
}

// LessionProgress 一节课程的完成进展
type LessionProgress struct {
	ID          bson.ObjectId   `bson:"_id" json:"id"`
	UserID      bson.ObjectId   `bson:"userID" json:"userID"`           //用户ID
	CourseID    bson.ObjectId   `bson:"courseID" json:"courseID"`       //课程ID
	Name        string          `bson:"name" json:"name"`               //课程名称
	Icon        string          `bson:"icon" json:"icon"`               //课程图标
	FinishItems []bson.ObjectId `bson:"finishItems" json:"finishItems"` //完成ID
	CurItemID   bson.ObjectId   `bson:"curItemID" json:"curItemID"`     //当前学习ID
}

// PostLessionProgress 完成一个视频提交进展
type PostLessionProgress struct {
	ID           bson.ObjectId `bson:"_id" json:"id"`
	UserID       bson.ObjectId `bson:"userID" json:"userID"`             //用户ID
	CourseID     bson.ObjectId `bson:"courseID" json:"courseID"`         //课程ID
	Name         string        `bson:"name" json:"name"`                 //课程名称
	Icon         string        `bson:"icon" json:"icon"`                 //课程图标
	FinishItemID bson.ObjectId `bson:"finishItemID" json:"finishItemID"` //完成ID
	CurItemID    bson.ObjectId `bson:"curItemID" json:"curItemID"`       //当前学习ID
}
type CourseProgress struct {
	TotalLessionCnt  int              `bson:"totalLessionCnt" json:"totalLessionCnt"`       //总课时数
	FishedLessionCnt int              `bson:"finishedLessionCnt" json:"finishedLessionCnt"` //完成的课时数
	LessionProgress  `bson:",inline"` //课程的完成情况
}
type AllCoursesProgress struct {
	CourseID         bson.ObjectId   `bson:"courseID" json:"courseID"` //课程ID
	Name             string          `bson:"name" json:"name"`         //课程名称
	Icon             string          `bson:"icon" json:"icon"`
	FinishItems      []bson.ObjectId `bson:"finishItems" json:"-"`                         //完成ID
	TotalLessionCnt  int             `bson:"totalLessionCnt" json:"totalLessionCnt"`       //总课时数
	FishedLessionCnt int             `bson:"finishedLessionCnt" json:"finishedLessionCnt"` //完成的课时数
}

// UpsertLessionProgress 插入或者更新一节课的学习进展
func (progMod *ProgressModels) UpsertLessionProgress(progress PostLessionProgress) error {
	var update bson.M
	logs.Info("progress:", progress)
	progMod.lessionPoint(progress)
	insertfun := func(col *mgo.Collection) error {
		query := bson.M{"userID": progress.UserID, "courseID": progress.CourseID}
		if progress.FinishItemID.Hex() == "" {
			update = bson.M{"$set": bson.M{"curItemID": progress.CurItemID, "name": progress.Name, "icon": progress.Icon}}
		} else {
			update = bson.M{"$addToSet": bson.M{"finishItems": progress.FinishItemID}, "$set": bson.M{"curItemID": progress.CurItemID}}
		}
		change := mgo.Change{
			Update:    update,
			Upsert:    true,
			ReturnNew: true,
		}
		_, err := col.Find(query).Apply(change, nil)
		return err
	}
	return progMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "courseprogress", insertfun)
}

// GetCourseProgress 获取课程的进展
func (progMod *ProgressModels) GetCourseProgress(userID string, courseID string) (*CourseProgress, error) {
	var progress LessionProgress
	var totalLessionCnt int
	f := func(col *mgo.Collection) error {
		return col.Find(bson.M{"courseID": bson.ObjectIdHex(courseID), "userID": bson.ObjectIdHex(userID)}).One(&progress)
	}
	err := progMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "courseprogress", f)
	if err != nil {
		return nil, err
	}
	courses, err := progMod.CourseMod.GetAllLessions(bson.ObjectIdHex(courseID))
	if err != nil {
		return nil, err
	}
	logs.Info("progress:", progress)
	courseProgress := new(CourseProgress)
	courseProgress.LessionProgress = progress
	courseProgress.FishedLessionCnt = len(progress.FinishItems)
	for _, lession := range courses.Lessions {
		totalLessionCnt += len(lession.Contents)
	}
	courseProgress.TotalLessionCnt = totalLessionCnt
	return courseProgress, nil
}

//GetAllCoursesProgress 获取该用户下的所有课程进展
func (progMod *ProgressModels) GetAllCoursesProgress(userID string) ([]AllCoursesProgress, error) {
	var allCoursesPros, coursesProgress []AllCoursesProgress
	var totalLessionCnt int
	f := func(col *mgo.Collection) error {
		return col.Find(bson.M{"userID": bson.ObjectIdHex(userID)}).Select(bson.M{"courseID": 1, "name": 1, "icon": 1, "finishItems": 1}).All(&allCoursesPros)
	}
	err := progMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "courseprogress", f)
	if err != nil {
		logs.Info("err:", err)
		return nil, err
	}
	for _, coursesPros := range allCoursesPros {
		totalLessionCnt = 0
		course, err := progMod.CourseMod.GetAllLessions(coursesPros.CourseID)
		if err != nil {
			return nil, err
		}
		for _, lession := range course.Lessions {
			totalLessionCnt += len(lession.Contents)
		}
		coursesPros.FishedLessionCnt = len(coursesPros.FinishItems)
		coursesPros.TotalLessionCnt = totalLessionCnt
		coursesProgress = append(coursesProgress, coursesPros)
	}
	return coursesProgress, err

}

//QueryStudentsProgress 查询指定课程学生学习情况
func (progMod *ProgressModels) QueryStudentsProgress(classCode, lessionID string) (interface{}, error) {
	result := []bson.M{}
	logs.Info("lessionID", lessionID)
	// query := bson.M{"finishItems": bson.ObjectIdHex(lessionID)}
	// f := func(col *mgo.Collection) error {
	// 	return col.Find(query).All(&result)
	// }
	pipeline := []bson.M{
		{"$match": bson.M{"code": classCode}},
		{"$unwind": "$students"},
		{"$lookup": bson.M{
			"from":         "courseprogress",
			"localField":   "students.userID",
			"foreignField": "userID",
			"as":           "learned",
		}},
		{
			"$match": bson.M{"learned.finishItems": bson.M{"$eq": bson.ObjectIdHex(lessionID)}},
		},
		{"$project": bson.M{
			"_id":             0,
			"students.userID": 1,
		}},
	}
	f := func(col *mgo.Collection) error {
		return col.Pipe(pipeline).All(&result)
	}
	err := progMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "classes", f)
	//err := progMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "courseprogress", f)
	return result, err
}

//LessionFinished 查询课程完成信息
// func LessionFinished(progress LessionProgress) (LessionProgress, error) {
// 	userID := progress.UserID
// 	var lessionProgress LessionProgress
// 	f := func(col *mgo.Collection) error {
// 		return col.Find(bson.M{"userID": userID, "lessionID": lessionID}).One(&lessionProgress)
// 	}
// 	err := progMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "courseprogress", f)
// 	return lessionProgress, err
// }

/*********************************************************************************************/
/*********************************** 以下为本控制器的内部函数 *********************************/
/*********************************** *********************************************************/
//lessionPoint 学完课程加一定积分
func (progMod *ProgressModels) lessionPoint(progress PostLessionProgress) {
	calculus := new(FinishedcouCalculus)
	if progress.FinishItemID.Hex() != "" {
		progMod.UserMod.UpdateUserCalculus(progress.UserID, calculus)
	}
}
