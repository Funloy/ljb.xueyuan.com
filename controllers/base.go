// @APIVersion 1.0.0
// @Title 控制器基类
// @Description 控制器基础类，为后面继承该类的控制器提供一些共用的方法或操作
// @Contact xuchuangxin@icanmake.cn
// @TermsOfServiceUrl https://maiyajia.com/
// @License
// @LicenseUrl

package controllers

import (
	"errors"
	"fmt"
	"runtime"
	"strconv"
	"strings"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
	m "maiyajia.com/models"
	"maiyajia.com/services/mongo"
	"maiyajia.com/services/token"
)

//BaseController 基础控制器
type (
	BaseController struct {
		beego.Controller
		mongo.MgoClient
	}
)

//NestPreparer 作为子类自定义prepare使用
type NestPreparer interface {
	NestPrepare()
}

//Prepare 作初始化处理
func (base *BaseController) Prepare() {
	base.MgoClient.StartSession()
	// 获取数据库
	if app, ok := base.AppController.(NestPreparer); ok {
		app.NestPrepare()
	}
}

//Finish 后处理
func (base *BaseController) Finish() {
	defer func() {
		base.MgoClient.CloseSession()
	}()
}

// checkToken 读取请求的token
func (base *BaseController) checkToken() *token.Token {
	// JWT验证
	token, err := getClientToken(base.Ctx.Input.Header("Authorization"))
	if err != nil {
		base.abortWithError(m.ERR_TOKEN_FMT_FAIL)
	}
	return token
}

// needAdminPermission 检查是否拥有管理员权限
func (base *BaseController) needAdminPermission(token *token.Token) {
	// 检查权限（角色为管理员）
	if token.UserRole != m.ROLE_ADMIN {
		base.abortWithError(m.ERR_PERMISSION_DENIED)
	}
	return
}

// needAdminOrTeacherPermission 检查是否拥有教师或管理员权限
func (base *BaseController) needAdminOrTeacherPermission(token *token.Token) {
	// 检查权限（角色为老师或管理员）
	if token.UserRole != m.ROLE_TEACHER && token.UserRole != m.ROLE_ADMIN {
		base.abortWithError(m.ERR_PERMISSION_DENIED)
	}
	return
}

// jsonResult 服务端返回json
func (base *BaseController) jsonResult(out interface{}) {
	base.Data["json"] = out
	base.ServeJSON()
}

// abortWithError 根据错误码获取错误描述信息，然后发送到请求客户端
func (base *BaseController) abortWithError(code int) {

	result := m.ErrorResult{
		Code:    code,
		Message: m.GetErrorMsgs(code),
	}
	base.Data["json"] = result
	base.ServeJSON()

	/* 调用 StopRun 之后，如果你还定义了 Finish 函数就不会再执行。
	* 如果需要释放资源，那么请自己在调用 StopRun 之前手工调用 Finish 函数。
	* https://beego.me/docs/mvc/controller/controller.md
	 */

	// 调用Finish()函数，释放数据库资源
	base.Finish()
	// 终止执行
	base.StopRun()
}

// getClientToken 根据客户端请求的令牌TOKEN字符串，解析出TOKEN信息
func getClientToken(authHeader string) (*token.Token, error) {

	auths := strings.Split(authHeader, " ")
	if len(auths) != 2 || auths[0] != "Bearer" {
		return nil, errors.New("authorization invalid")
	}

	token, err := token.ParseToken(auths[1])
	if err != nil {
		return nil, errors.New("authorization invalid")
	}
	return token, nil
}

/**
* 解析请求中的number和page参数，返回Pageing类型
 */
func paramPaging(context *context.Context) (*m.Paging, error) {
	number, err := strconv.Atoi(context.Input.Param(":number"))
	if err != nil {
		return nil, errors.New("number is invalid") // 参数解析错误
	}
	// 请求数量(number)的值不能大于或小于系统规定的最大或最小请求数量
	if number > m.MaxQueryNumber || number < m.MinQueryNumber {
		return nil, errors.New("number is out of range")
	}

	page, err := strconv.Atoi(context.Input.Param(":page"))
	if err != nil {
		return nil, errors.New("page is invalid")
	}

	// page的值不能小于系统规定的最小请求数量
	if page < m.MinQueryPage {
		return nil, errors.New("page is out of range")
	}

	return m.NewPaging(page, number), nil
}

//Recover 宕机时处理函数
func (base *BaseController) Recover(code int) {
	// 发生宕机时，获取panic传递的上下文并打印
	err := recover()
	switch err.(type) {
	case runtime.Error: // 运行时错误
		base.abortWithError(code)
	default: // 非运行时错误
		fmt.Println("error:", err)
	}
}
