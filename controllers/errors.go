// @APIVersion 1.0.0
// @Title 接口请求错误处理控制器
// @Description 本控制器提供HTTP请求时发生的401、403、404、500、503 这几种错误的处理
// @Contact xuchuangxin@icanmake.cn
// @TermsOfServiceUrl https://maiyajia.com/
// @License
// @LicenseUrl

package controllers

import (
	"net/http"

	"github.com/astaxie/beego"
	"maiyajia.com/models"
)

// ErrorController 错误处理控制器
type ErrorController struct {
	beego.Controller
}

// Error404 404错误返回
func (ec *ErrorController) Error404() {
	ec.Data["json"] = &models.ErrorResult{Code: http.StatusNotFound, Message: "api not found"}
	ec.ServeJSON()
}
