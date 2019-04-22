// @Title 用户学习进度控制器
// @Description 用户学习进度控制器

package controllers

import (
	"encoding/json"

	"github.com/astaxie/beego/logs"
	"gopkg.in/mgo.v2/bson"
	"maiyajia.com/models"

	m "maiyajia.com/models"
)

// ProgressController 学习进度相关的控制器
type ProgressController struct {
	BaseController
	ProgMod m.ProgressModels
}

// NestPrepare 初始化数据库
func (progressCtrl *ProgressController) NestPrepare() {
	progressCtrl.ProgMod.MgoSession = &progressCtrl.MgoClient
	progressCtrl.ProgMod.CourseMod.MgoSession = &progressCtrl.MgoClient
	progressCtrl.ProgMod.UserMod.MgoSession = progressCtrl.MgoClient
}

// UploadLessProgress 上传课节进度
func (progressCtrl *ProgressController) UploadLessProgress() {
	// 获取token
	token := progressCtrl.checkToken()
	var progress models.PostLessionProgress
	if err := json.Unmarshal(progressCtrl.Ctx.Input.RequestBody, &progress); err != nil {
		progressCtrl.abortWithError(m.ERR_REQUEST_PARAM)
	}
	logs.Info("progress:", progress)
	if progress.ID.Hex() == "" {
		progress.ID = bson.NewObjectId()
		progress.UserID = bson.ObjectIdHex(token.UserID)
	}
	if err := progressCtrl.ProgMod.UpsertLessionProgress(progress); err != nil {
		progressCtrl.abortWithError(m.ERR_LESSION_PROGRESS_UPDATE_FAIL)
	}
	// 封装返回数据
	out := make(map[string]interface{})
	out["code"] = 0
	progressCtrl.jsonResult(out)
}

// GetCourseProgress 查询一个课程下每个课时的进展以及总进展
func (progressCtrl *ProgressController) GetCourseProgress() {
	var (
		courseID string
	)
	// 获取token
	token := progressCtrl.checkToken()
	out := make(map[string]interface{})
	if err := progressCtrl.Ctx.Input.Bind(&courseID, "courseid"); err != nil {
		progressCtrl.abortWithError(m.ERR_REQUEST_PARAM)
	}
	if courseID == "" {
		courseProgress, err := progressCtrl.ProgMod.GetAllCoursesProgress(token.UserID)
		if err != nil && "not found" != err.Error() {
			progressCtrl.abortWithError(m.ERR_LESSION_PROGRESS_UPDATE_FAIL)
		}
		out["courseprogress"] = courseProgress
	} else {
		courseProgress, err := progressCtrl.ProgMod.GetCourseProgress(token.UserID, courseID)
		if err != nil && "not found" != err.Error() {
			logs.Info("err:", err)
			progressCtrl.abortWithError(m.ERR_LESSION_PROGRESS_UPDATE_FAIL)
		}
		out["courseprogress"] = courseProgress
	}
	out["code"] = 0
	progressCtrl.jsonResult(out)

}

//GetStudnetsProgress 获取指定课程下的学生学习进度
func (progressCtrl *ProgressController) GetStudentsProgress() {
	// token := progressCtrl.checkToken()
	// progressCtrl.needAdminOrTeacherPermission(token)
	var lessonID string
	var classCode string

	if err := progressCtrl.Ctx.Input.Bind(&lessonID, "lessonID"); err != nil {
		progressCtrl.abortWithError(m.ERR_REQUEST_PARAM)
	}
	if err := progressCtrl.Ctx.Input.Bind(&classCode, "classCode"); err != nil {
		progressCtrl.abortWithError(m.ERR_REQUEST_PARAM)
	}
	result, _ := progressCtrl.ProgMod.QueryStudentsProgress(classCode, lessonID)
	// 封装返回数据
	out := make(map[string]interface{})
	out["code"] = 0
	out["result"] = result
	progressCtrl.jsonResult(out)

}
