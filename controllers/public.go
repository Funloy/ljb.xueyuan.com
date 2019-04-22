// @APIVersion 1.0.0
// @Title 公共接口服务
// @Description 用于无需登录就能使用的公共API接口服务
// @Contact xuchuangxin@icanmake.cn
// @TermsOfServiceUrl https://maiyajia.com/
// @License
// @LicenseUrl

package controllers

import (
	"encoding/json"

	"gopkg.in/mgo.v2/bson"
	m "maiyajia.com/models"
	"maiyajia.com/services/avatar"
	"maiyajia.com/util"
)

// PublicController 控制器
type PublicController struct {
	BaseController
	userMod  m.UserModels
	classMod m.ClassModels
	calMod   m.CalModels
}

// NestPrepare 初始化函数
// 把控制器的MgoClient赋值到模型的数据库操作客户端
func (pubCtl *PublicController) NestPrepare() {
	pubCtl.userMod.MgoSession = pubCtl.MgoClient
	pubCtl.classMod.MgoSession = &pubCtl.MgoClient
	pubCtl.calMod.MgoSession = &pubCtl.MgoClient
}

// CheckToken 查询用户登陆是否过期
func (pubCtl *PublicController) CheckToken() {
	// 获取token
	pubCtl.checkToken()
	// 封装返回数据
	out := make(map[string]interface{})
	out["code"] = 0
	pubCtl.jsonResult(out)
}

// QueryUser 查询用户信息接口
func (pubCtl *PublicController) QueryUser() {
	uid := pubCtl.Ctx.Input.Param(":userid")
	if !bson.IsObjectIdHex(uid) {
		pubCtl.abortWithError(m.ERR_REQUEST_PARAM)
	}
	userid := bson.ObjectIdHex(uid)
	user, err := pubCtl.userMod.FetchUserProfile(userid)
	if err != nil {
		pubCtl.abortWithError(m.ERR_USER_NONE)
	}

	// 封装返回数据
	out := make(map[string]interface{})
	out["code"] = 0
	out["user"] = user
	pubCtl.jsonResult(out)
}

// RandomAvatar 随机返回系统的头像图片. 注：本接口提供给用户修改更新头像使用
func (pubCtl *PublicController) RandomAvatar() {
	// 随机产生性别和salt字符串
	var gender avatar.Gender
	if util.RandomInt(0, 2) == 1 {
		gender = avatar.FEMALE
	} else {
		gender = avatar.MALE
	}
	var salt = util.CreateRandomString(8)
	img, _ := avatar.BuilderWithSalt(gender, salt)
	imgSrc := avatar.ImageToBase64(img)

	// 封装返回数据
	out := make(map[string]interface{})
	out["code"] = 0
	out["gender"] = gender
	out["charset"] = salt
	out["src"] = imgSrc
	pubCtl.jsonResult(out)

}

// QueryClass 查询班级
func (pubCtl *PublicController) QueryClass() {

	// 获取GET请求参数（注：请求参数放置在路由url中，请参见router.go）
	classCode := pubCtl.Ctx.Input.Param(":code")

	//查询班级
	class, err := pubCtl.classMod.FetchClassDetails(classCode)
	if err != nil {
		pubCtl.abortWithError(m.ERR_CLASS_NONE)
	}
	// 封装返回数据
	out := make(map[string]interface{})
	out["code"] = 0
	out["class"] = class

	pubCtl.jsonResult(out)
}

// QueryJoinClasses 查询用户已经加入的班级
func (pubCtl *PublicController) QueryJoinClasses() {

	// 获取GET请求参数（注：请求参数放置在路由url中，请参见router.go）
	username := pubCtl.Ctx.Input.Param(":username")

	classes, err := pubCtl.userMod.FetchJoinClasses(username)
	if err != nil {
		pubCtl.abortWithError(m.ERR_USER_NONE)
	}
	// 封装返回数据
	out := make(map[string]interface{})
	out["code"] = 0
	out["joinClasses"] = classes
	pubCtl.jsonResult(out)
}

// QueryCreatedClasses 查询用户已创建的班级列表
func (pubCtl *PublicController) QueryCreatedClasses() {
	id := pubCtl.Ctx.Input.Param(":id")
	if !bson.IsObjectIdHex(id) {
		pubCtl.abortWithError(m.ERR_REQUEST_PARAM)
	}
	classes, err := pubCtl.classMod.ClassesByCreator(bson.ObjectIdHex(id))
	if err != nil {
		pubCtl.abortWithError(m.ERR_USER_NONE)
	}
	// 封装返回数据
	out := make(map[string]interface{})
	out["code"] = 0
	out["classes"] = classes
	pubCtl.jsonResult(out)

}

// GetAllMedals 查询勋章
func (pubCtl *PublicController) GetAllMedals() {

	medals := m.FetchAllMedals()
	out := make(map[string]interface{})
	out["code"] = 0
	out["result"] = medals
	pubCtl.jsonResult(out)
}

// GetCalculusRank 用户积分排行榜
func (pubCtl *PublicController) GetCalculusRank() {
	result, err := pubCtl.calMod.CalculusRank()
	if err != nil {
		pubCtl.abortWithError(m.ERR_CALCULUS_RANK_QUERY_FAIL)
	}
	// 封装返回数据
	out := make(map[string]interface{})
	out["code"] = 0
	out["rank"] = result
	pubCtl.jsonResult(out)
}

// GetClassCalculusRank 用户积分排行榜
func (pubCtl *PublicController) GetClassCalculusRank() {
	code := pubCtl.Ctx.Input.Param(":code")
	result, err := pubCtl.calMod.CalculusRankInClass(code)
	if err != nil {
		pubCtl.abortWithError(m.ERR_CALCULUS_RANK_QUERY_FAIL)
	}
	// 封装返回数据
	out := make(map[string]interface{})
	out["code"] = 0
	out["rank"] = result
	pubCtl.jsonResult(out)
}

// SubmitApplication 用户提交成为教师申请
func (pubCtl *PublicController) SubmitApplication() {
	var apply m.Apply
	// 获取token
	token := pubCtl.checkToken()
	if err := json.Unmarshal(pubCtl.Ctx.Input.RequestBody, &apply); err != nil {
		pubCtl.abortWithError(m.ERR_REQUEST_PARAM)
	}
	userid := token.UserID
	username := token.Username
	// if qualify, err := m.QueryApplication(userid); err == nil {
	// 	if qualify.Status == 2 {
	// 		pubCtl.abortWithError(m.ERR_USER_APPLICATION_SUBMIT_EXSIT)
	// 	} else if qualify.Status == 1 {
	// 		pubCtl.abortWithError(m.ERR_USER_APPLICATION_SUCCESS)
	// 	}
	// }
	if err := pubCtl.userMod.NewApplication(userid, username, apply); err != nil {
		pubCtl.abortWithError(m.ERR_USER_APPLICATION_SUBMIT_FAIL)
	}
	// 封装返回数据
	out := make(map[string]interface{})
	out["code"] = 0
	pubCtl.jsonResult(out)
}

// GetApplicationProgress 用户查看申请进度
func (pubCtl *PublicController) GetApplicationProgress() {
	// 获取token
	token := pubCtl.checkToken()
	qualify, err := pubCtl.userMod.QueryApplication(token.UserID)
	if err != nil {
		pubCtl.abortWithError(m.ERR_USER_APPLICATION_PROGRESS_FAIL)
	}
	// 封装返回数据
	out := make(map[string]interface{})
	out["code"] = 0
	out["qualify"] = qualify
	pubCtl.jsonResult(out)
}

/*************  一下为测试接口， 发布版本是需要删除 *************/
