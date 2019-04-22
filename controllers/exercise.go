package controllers

import (
	"encoding/json"

	"gopkg.in/mgo.v2/bson"
	m "maiyajia.com/models"
)

// ExercisesController 习题控制器
type ExercisesController struct {
	BaseController
	exerMod m.ExerModels
}

// NestPrepare 数据库客户端
func (exerCtrl *ExercisesController) NestPrepare() {
	exerCtrl.exerMod.MgoSession = &exerCtrl.MgoClient
}

//SaveExer 保存练习信息
func (exerCtrl *ExercisesController) SaveExer() {
	userID := exerCtrl.checkToken().UserID
	var exer m.Exercise
	if err := json.Unmarshal(exerCtrl.Ctx.Input.RequestBody, &exer); err != nil {
		exerCtrl.abortWithError(m.ERR_REQUEST_PARAM)

	}
	if exer.ID.Hex() == "" {
		exer.ID = bson.NewObjectId()
		exer.UserID = bson.ObjectIdHex(userID)
	}
	if err := exerCtrl.exerMod.InsertExercise(exer); err != nil {
		exerCtrl.abortWithError(m.ERR_ADD_EXER_FAIL)
	}
	out := make(map[string]interface{})
	out["code"] = 0
	exerCtrl.jsonResult(out)
}
