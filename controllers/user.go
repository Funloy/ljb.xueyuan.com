// @APIVersion 1.0.0
// @Title 用户接口服务
// @Description 用户账号信息和配置管理的接口服务
// @Contact xuchuangxin@icanmake.cn
// @TermsOfServiceUrl https://maiyajia.com/
// @License
// @LicenseUrl

package controllers

import (
	"encoding/json"

	"gopkg.in/mgo.v2/bson"

	m "maiyajia.com/models"
)

// UserController 登录注册控制器
type UserController struct {
	BaseController
	userMod    m.UserModels
	messageMod m.MessageModels
	classMod   m.ClassModels
}

type updateAvatarCredential struct {
	Gender  int    `form:"gender"`  // 用户选择头像性别
	Charset string `form:"charset"` // 用户选择头像图片对应的编码
}

// joinClassCredential 用户加入班级请求凭证
type joinClassCredential struct {
	ClassCode string `form:"classCode"`
}

// classCredential 用户退出班级请求凭证
type quitClassCredential struct {
	ClassCode string `form:"classCode"`
}

// readMessageCredential 阅读消息请求凭证
type readMessageCredential struct {
	MessageID string `form:"messageID"`
}

// NestPrepare 初始化函数
// 把控制器的MgoClient赋值到模型的数据库操作客户端
func (userCtl *UserController) NestPrepare() {
	userCtl.userMod.MgoSession = userCtl.MgoClient
	userCtl.messageMod.MgoSession = &userCtl.MgoClient
	userCtl.classMod.MgoSession = &userCtl.MgoClient
}

// UpdateAvatar 更新头像
func (userCtl *UserController) UpdateAvatar() {
	// 获取token
	token := userCtl.checkToken()
	var credential updateAvatarCredential
	// 凭证解析错误
	if err := json.Unmarshal(userCtl.Ctx.Input.RequestBody, &credential); err != nil {
		userCtl.abortWithError(m.ERR_REQUEST_PARAM)
	}

	avatar, err := userCtl.userMod.UpdateAvatar(token.Username, credential.Gender, credential.Charset)
	if err != nil {
		userCtl.abortWithError(m.ERR_AVATAR_CHAN_FAIL)
	}

	// 封装返回数据
	out := make(map[string]interface{})
	out["code"] = 0
	out["avatar"] = avatar
	userCtl.jsonResult(out)

}

// JoinClass 学生加入班级
func (userCtl *UserController) JoinClass() {
	// 获取token
	token := userCtl.checkToken()

	var credential joinClassCredential

	// 凭证解析错误
	if err := json.Unmarshal(userCtl.Ctx.Input.RequestBody, &credential); err != nil {
		userCtl.abortWithError(m.ERR_REQUEST_PARAM)
	}

	// 检查用户是不是该班级的成员
	ok := userCtl.userMod.IsJoinedClass(token.Username, credential.ClassCode)
	if ok {
		userCtl.abortWithError(m.ERR_CLASS_USER_EXISTS)
	}

	// 加入班级
	if err := userCtl.classMod.AddStudentToClass(bson.ObjectIdHex(token.UserID), credential.ClassCode); err != nil {
		userCtl.abortWithError(m.ERR_CLASS_JOIN_FAIL)
	}

	class, _ := userCtl.classMod.FetchClassDetails(credential.ClassCode)
	// 封装返回数据
	out := make(map[string]interface{})
	out["code"] = 0
	out["class"] = class
	userCtl.jsonResult(out)
}

// QuitClass 用户退出班级
func (userCtl *UserController) QuitClass() {
	// 获取token
	token := userCtl.checkToken()

	var credential joinClassCredential

	// 凭证解析错误
	if err := json.Unmarshal(userCtl.Ctx.Input.RequestBody, &credential); err != nil {
		userCtl.abortWithError(m.ERR_REQUEST_PARAM)
	}

	// 检查用户是不是该班级的成员
	ok := userCtl.userMod.IsJoinedClass(token.Username, credential.ClassCode)
	if !ok {
		userCtl.abortWithError(m.ERR_CLASS_USER_NONE)
	}

	// 退出班级
	if err := userCtl.classMod.RemoveStudentFromClass(bson.ObjectIdHex(token.UserID), credential.ClassCode); err != nil {
		userCtl.abortWithError(m.ERR_CLASS_QUIT_FAIL)
	}
	// 封装返回数据
	out := make(map[string]interface{})
	out["code"] = 0
	userCtl.jsonResult(out)
}

// UnreadMessageCount 查询用户的未读消息数目
func (userCtl *UserController) UnreadMessageCount() {
	// 获取token
	token := userCtl.checkToken()

	count, err := userCtl.messageMod.QueryUnreadMessageCount(bson.ObjectIdHex(token.UserID))
	if err != nil {
		userCtl.abortWithError(m.ERR_REQUEST_PARAM)
	}

	out := make(map[string]interface{})
	out["code"] = 0
	out["unread"] = count
	userCtl.jsonResult(out)
}

// GetMessageCount 获取消息的数量信息
func (userCtl *UserController) GetMessageCount() {
	// 获取token
	token := userCtl.checkToken()

	count, _ := userCtl.messageMod.QueryMessageCount(bson.ObjectIdHex(token.UserID))
	out := make(map[string]interface{})
	out["code"] = 0
	out["count"] = count
	userCtl.jsonResult(out)
}

// ReadNewMessage 用户阅读新的消息
func (userCtl *UserController) ReadNewMessage() {
	// 获取token
	token := userCtl.checkToken()

	// 凭证解析错误
	var credential readMessageCredential
	if err := json.Unmarshal(userCtl.Ctx.Input.RequestBody, &credential); err != nil {
		userCtl.abortWithError(m.ERR_REQUEST_PARAM)
	}

	err := userCtl.messageMod.ReadMessage(bson.ObjectIdHex(token.UserID), bson.ObjectIdHex(credential.MessageID))
	if err != nil {
		userCtl.abortWithError(m.ERR_CLASS_MESSAGE_READ_FAIL)
	}

	// 封装返回数据
	out := make(map[string]interface{})
	out["code"] = 0
	userCtl.jsonResult(out)

}

// QueryMessages 查询用户消息列表
func (userCtl *UserController) QueryMessages() {
	// 获取token
	token := userCtl.checkToken()

	// 获取请求参数中的分页数据
	paging, err := paramPaging(userCtl.Ctx)
	if err != nil {
		userCtl.abortWithError(m.ERR_REQUEST_PARAM)
	}

	// 拉取用户订阅的消息
	messages, err := userCtl.messageMod.FetchMessages(bson.ObjectIdHex(token.UserID), paging)
	if err != nil {
		userCtl.abortWithError(m.ERR_CLASS_MESSAGE_QUERY_FAIL)
	}

	// 封装返回数据
	out := make(map[string]interface{})
	out["code"] = 0
	out["messages"] = messages
	userCtl.jsonResult(out)

}

// DeleteMessages 删除用户消息
func (userCtl *UserController) DeleteMessages() {
	// 获取token
	token := userCtl.checkToken()
	var messageID string
	if userCtl.Ctx.Input.Bind(&messageID, "messageID") != nil {
		userCtl.abortWithError(m.ERR_REQUEST_PARAM)
	}
	if userCtl.messageMod.DeleteUserMessage(bson.ObjectIdHex(token.UserID), bson.ObjectIdHex(messageID)) != nil {
		userCtl.abortWithError(m.ERR_CLASS_MESSAGE_DELETE_FAIL)
	}
	// 封装返回数据
	out := make(map[string]interface{})
	out["code"] = 0
	userCtl.jsonResult(out)

}
