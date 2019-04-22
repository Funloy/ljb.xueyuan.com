// @APIVersion 1.0.0
// @Title 班级管理接口服务，主要用户教师或管理员对班级进行管理
// @Description 班级信息和配置管理的接口服务
// @Contact xuchuangxin@icanmake.cn
// @TermsOfServiceUrl https://maiyajia.com/
// @License
// @LicenseUrl

package controllers

import (
	"encoding/json"
	"fmt"
	"regexp"

	"gopkg.in/mgo.v2/bson"
	m "maiyajia.com/models"
)

// ClassController 班级管理控制器
type ClassController struct {
	BaseController
	userMod    m.UserModels
	classMod   m.ClassModels
	messageMod m.MessageModels
}

// NestPrepare 初始化函数
// 把控制器的MgoClient赋值到模型的数据库操作客户端
func (classCtl *ClassController) NestPrepare() {
	classCtl.userMod.MgoSession = classCtl.MgoClient
	classCtl.classMod.MgoSession = &classCtl.MgoClient
	classCtl.classMod.UserMod.MgoSession = classCtl.MgoClient
	classCtl.messageMod.MgoSession = &classCtl.MgoClient
}

// classCredential 班级管理接口请求凭证
type classCredential struct {
	ClassID   string `form:"classID"`
	UserID    string `form:"userID"`
	ClassName string `form:"className"`
	ClassCode string `form:"classCode"`
}

// inviteCredential 邀请学生加入班级的请求参数
type inviteCredential struct {
	Username  string `form:"username"`  // 邀请的学生的用户名
	ClassCode string `form:"classCode"` // 班级的代码
}

//  withdrawCredential 把学生从班级移除的请求参数
type withdrawCredential struct {
	Username  string `form:"username"`  // 被移除学生的用户名
	ClassCode string `form:"classCode"` // 班级的代码
}

// destroyCredential 注销班级请求参数
type destroyCredential struct {
	ClassCode string `form:"classCode"` // 班级的代码
}

// messageCredentials 发布消息请求凭证
// type publishMessageCredentials struct {
// 	ClassID string `form:"classID"`
// 	Title   string `form:"title"`
// 	Content string `form:"Content"`
// }
type publishMessageCredentials struct {
	ClassID   string `json:"classID"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	ClassName string `json:"className"`
	Logo      string `json:"logo"`
}

// RegisterClass 注册一个新的班级
func (classCtl *ClassController) RegisterClass() {

	// 获取token
	token := classCtl.checkToken()

	// 检查是否拥有创建班级的权限（角色为老师或管理员）
	classCtl.needAdminOrTeacherPermission(token)

	var credential classCredential
	// 凭证解析错误
	if err := json.Unmarshal(classCtl.Ctx.Input.RequestBody, &credential); err != nil {
		classCtl.abortWithError(m.ERR_REQUEST_PARAM)
	}
	// 验证班级名称是否符合要求
	if !validateClassName(credential.ClassName) {
		classCtl.abortWithError(m.ERR_CLASS_FMT_FAIL)
	}
	//查找班级名是否被占有
	if classCtl.classMod.HasClassName(credential.ClassName) {
		classCtl.abortWithError(m.ERR_CLASS_EXISTS)
	}

	class := classCtl.classMod.NewClass(bson.ObjectIdHex(token.UserID), credential.ClassName)
	if err := classCtl.classMod.CreateClass(class); err != nil {
		classCtl.abortWithError(m.ERR_CLASS_CREATE_FAIL)
	}

	// 封装返回数据
	out := make(map[string]interface{})
	out["code"] = 0
	out["class"] = class

	classCtl.jsonResult(out)
}

// RenameClass 更改班级名称，
// 本接口用于教师更改班级名称
// 注意：必须满足2个条件：(1)新名称的唯一性；(2)操作者就是该班级的创建者(教师或管理员）
func (classCtl *ClassController) RenameClass() {
	// 获取token
	token := classCtl.checkToken()
	// 检查权限（角色为老师或管理员）
	classCtl.needAdminOrTeacherPermission(token)

	// 凭证解析错误
	var credential classCredential
	if err := json.Unmarshal(classCtl.Ctx.Input.RequestBody, &credential); err != nil {
		classCtl.abortWithError(m.ERR_REQUEST_PARAM)
	}

	// 检查班级代码
	if !validateClassCode(credential.ClassCode) {
		classCtl.abortWithError(m.ERR_CLASS_CODE_FAIL)
	}

	// 检查新的班级名称，新的名称是否唯一性
	if !validateClassName(credential.ClassName) {
		classCtl.abortWithError(m.ERR_CLASS_FMT_FAIL)
	}
	if classCtl.classMod.HasClassName(credential.ClassName) {
		classCtl.abortWithError(m.ERR_CLASS_EXISTS)
	}

	// 更新班级名称
	if err := classCtl.classMod.RenameClassName(token.UserID, credential.ClassCode, credential.ClassName); err != nil {
		classCtl.abortWithError(m.ERR_CLASS_RENAME_FAIL)
	}

	// 封装返回数据
	out := make(map[string]interface{})
	out["code"] = 0
	classCtl.jsonResult(out)

}

// Invite 邀请学生加入班级
// 本接口用于教师添加学生
func (classCtl *ClassController) Invite() {
	// 获取token
	token := classCtl.checkToken()

	// 检查权限（角色为老师或管理员）
	classCtl.needAdminOrTeacherPermission(token)

	var credential inviteCredential

	// 凭证解析错误
	if err := json.Unmarshal(classCtl.Ctx.Input.RequestBody, &credential); err != nil {
		classCtl.abortWithError(m.ERR_REQUEST_PARAM)
	}

	// 检查班级代码格式
	if !validateClassCode(credential.ClassCode) {
		classCtl.abortWithError(m.ERR_CLASS_CODE_FAIL)
	}

	// 检查用户是否存在
	user, err := classCtl.userMod.FindUserByUsername(credential.Username)
	if err != nil {
		classCtl.abortWithError(m.ERR_USER_NONE)
	}

	// 加入班级
	if err := classCtl.classMod.AddStudentToClass(user.ID, credential.ClassCode); err != nil {
		classCtl.abortWithError(m.ERR_CLASS_USER_EXISTS)
	}
	// 封装返回数据
	out := make(map[string]interface{})
	out["code"] = 0
	out["user"] = user
	classCtl.jsonResult(out)
}

// KickOut 把学生移出班级
// 本接口用于教师移除班级里的学生
func (classCtl *ClassController) KickOut() {

	token := classCtl.checkToken()

	// 检查权限（角色为老师或管理员）
	classCtl.needAdminOrTeacherPermission(token)

	var credential withdrawCredential
	// 凭证解析错误
	if err := json.Unmarshal(classCtl.Ctx.Input.RequestBody, &credential); err != nil {
		classCtl.abortWithError(m.ERR_REQUEST_PARAM)
	}

	// 检查班级代码格式
	if !validateClassCode(credential.ClassCode) {
		classCtl.abortWithError(m.ERR_CLASS_CODE_FAIL)
	}

	// 检查用户是否存在
	user, err := classCtl.userMod.FindUserByUsername(credential.Username)
	if err != nil {
		classCtl.abortWithError(m.ERR_USER_NONE)
	}

	// 从班级里移除学生
	if err := classCtl.classMod.RemoveStudentFromClass(user.ID, credential.ClassCode); err != nil {
		classCtl.abortWithError(m.ERR_CLASS_KICK_FAIL)
	}
	// 封装返回数据
	out := make(map[string]interface{})
	out["code"] = 0
	classCtl.jsonResult(out)

}

// DestroyClass 邀请学生加入班级
// 本接口用于教师注销班级。 注：注销班级的前提是，该班级的学生人数为0
func (classCtl *ClassController) DestroyClass() {

	token := classCtl.checkToken()

	// 检查权限（角色为老师或管理员）
	classCtl.needAdminOrTeacherPermission(token)

	var credential destroyCredential
	// 凭证解析错误
	if err := json.Unmarshal(classCtl.Ctx.Input.RequestBody, &credential); err != nil {
		classCtl.abortWithError(m.ERR_REQUEST_PARAM)
	}

	// 查询班级的人数是否为0，为学生人数为0才符合被注销的条件
	class, err := classCtl.classMod.FindClassByCode(credential.ClassCode)
	if err != nil {
		classCtl.abortWithError(m.ERR_CLASS_CODE_FAIL)
	}
	if len(class.Students) != 0 {
		classCtl.abortWithError(m.ERR_CLASS_DESTROY_NOBODY)
	}

	if err := classCtl.classMod.DeleteClass(credential.ClassCode); err != nil {
		classCtl.abortWithError(m.ERR_CLASS_DESTROY_FAIL)
	}
	// 封装返回数据
	out := make(map[string]interface{})
	out["code"] = 0
	classCtl.jsonResult(out)

}

// PublishMessage 发布班级消息
func (classCtl *ClassController) PublishMessage() {
	// 获取token
	token := classCtl.checkToken()
	// 检查权限（角色为老师或管理员）
	classCtl.needAdminOrTeacherPermission(token)
	// 凭证解析错误
	var credential publishMessageCredentials
	if err := json.Unmarshal(classCtl.Ctx.Input.RequestBody, &credential); err != nil {
		classCtl.abortWithError(m.ERR_REQUEST_PARAM)
	}
	// 发布消息
	err := classCtl.messageMod.PublishClassMessage(credential.Title, credential.Content, credential.ClassName, credential.Logo, bson.ObjectIdHex(credential.ClassID), bson.ObjectIdHex(token.UserID))
	if err != nil && "not student" == err.Error() {
		classCtl.abortWithError(m.ERR_CLASS_MESSAGE_PUBLISH)
	} else if err != nil {
		classCtl.abortWithError(m.ERR_CLASS_MESSAGE_PUBLISH_FAIL)
	}
	// 封装返回数据
	out := make(map[string]interface{})
	out["code"] = 0
	classCtl.jsonResult(out)

}

// @Title 班级名称验证
// @Description 班级名验证，长度为4~64位
// @Param	name	string	true	"班级名"
// @Success true {bool}
// @Failure false {bool}
func validateClassName(name string) bool {
	// //用户名正则，双字节字符（包括汉字），长度在4~64之间
	//reg := regexp.MustCompile(`^[a-z0-9A-Z\p{Han}]+(_[a-z0-9A-Z\p{Han}]+)*$`)
	if len(name) < 4 || len(name) > 64 {
		return false
	}
	reg := regexp.MustCompile(`[^\x00-\xff]`)
	return reg.MatchString(name)
}

// @Title 班级代码验证
// @Description 班级代码验证，长度为models.CLASS_CODE_LEN定义
// @Param	code	string	true	"班级代码"
// @Success true {bool}
// @Failure false {bool}
func validateClassCode(code string) bool {
	// //班级代码正则，0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ中的字符，长度为models.CLASS_CODE_LEN定义
	pattern := fmt.Sprintf(`^[0-9A-Z]{%d}$`, m.CLASS_CODE_LEN)
	reg := regexp.MustCompile(pattern)
	return reg.MatchString(code)
}
