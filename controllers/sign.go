// @APIVersion 1.0.0
// @Title 用户签到服务
// @Description 用户签到和签到查询接口服务
// @Contact xuchuangxin@icanmake.cn
// @TermsOfServiceUrl https://maiyajia.com/
// @License
// @LicenseUrl

package controllers

import (
	"strconv"
	"time"

	"gopkg.in/mgo.v2/bson"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"

	m "maiyajia.com/models"
)

// SignController 登录注册控制器
type SignController struct {
	BaseController
	userMod m.UserModels
	signMod m.SignModels
}

// NestPrepare 初始化函数
// 把控制器的MgoClient赋值到模型的数据库操作客户端
func (signCtl *SignController) NestPrepare() {
	signCtl.userMod.MgoSession = signCtl.MgoClient
	signCtl.signMod.MgoSession = &signCtl.MgoClient
}

// signedRequest 签到记录请求字段
type signedRequest struct {
	Year  int `form:"year"`
	Month int `form:"month"`
}

// Signed 用户签到
func (signCtl *SignController) Signed() {
	// 获取token
	token := signCtl.checkToken()
	uid := bson.ObjectIdHex(token.UserID)

	signed, status := signCtl.signMod.SignedLogger(uid)
	// 今日已经签到
	if !status {
		signCtl.abortWithError(m.ERR_SIGN_DUP_FAIL)
	}
	// 返回的signed为nil时，表示签到错误(内部系统错误)
	if signed == nil {
		signCtl.abortWithError(m.ERR_SIGN_FAIL)
	}

	// 获取签到微积分
	calculus := new(m.SignCalculus)
	calculus.SetDays(signed.ContinuousSigned)
	signCtl.userMod.UpdateUserCalculus(uid, calculus)

	// 是否能够荣获签到勋章，注:签到勋章一般都要求连续签到2天以上
	var signMedal *m.Medal
	if signed.ContinuousSigned > 1 {
		medalInfo := new(m.SignMedal)
		medalInfo.SetDays(signed.ContinuousSigned)
		if medal := signCtl.userMod.GrantUserMedal(uid, medalInfo); medal != nil {
			signMedal = medal
		}
	}

	// 签到成功，返回结果
	out := make(map[string]interface{})
	out["code"] = 0
	out["signed"] = signed
	out["calculus"] = calculus.Calculus()
	out["medal"] = signMedal
	signCtl.jsonResult(out)
}

// SignedRecord 查询签到记录
func (signCtl *SignController) SignedRecord() {

	// 获取token
	token := signCtl.checkToken()

	// 获取GET请求参数（注：请求参数放置在路由url中，已经为year和month参数设置了正则匹配，必须为整数）
	year, _ := strconv.Atoi(signCtl.Ctx.Input.Param(":year"))
	month, _ := strconv.Atoi(signCtl.Ctx.Input.Param(":month"))

	// 根据年月参数获取该年月份的起始日期和终止日期
	theMon := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	//monStart := theMon.AddDate(0, 0, 0).Format("2006-01-02 15:04:05")
	// monEnd := theMon.AddDate(0, 1, -1).Format("2006-01-02 15:04:05")
	monStart := theMon.AddDate(0, 0, 0)
	monEnd := theMon.AddDate(0, 1, -1)
	logs.Info("monStart,monEnd:", monStart, monEnd)
	signedRecords, err := signCtl.signMod.SignedRecords(token.UserID, monStart, monEnd)
	if err != nil {
		beego.Error(err)
		signCtl.abortWithError(m.ERR_SIGN_RECORD_FAIL)
	}
	out := make(map[string]interface{})
	out["code"] = 0
	out["signed"] = signedRecords

	// 返回结果
	signCtl.jsonResult(out)

}
