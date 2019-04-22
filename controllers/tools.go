//@Descriptio: 工具管理控制器
//@Autho: liaojiebiao
package controllers

import (
	"net/http"
	"os"
	"path"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/gorilla/websocket"
	m "maiyajia.com/models"
	"maiyajia.com/services/daemon"
)

// ToolsController operations for Tools
type ToolsController struct {
	BaseController
	ToolMod       m.ToolModels
	upgradeClient daemon.UpgradeModels
}

func (toolsCtrl *ToolsController) NestPrepare() {
	toolsCtrl.ToolMod.MgoSession = &toolsCtrl.MgoClient
	toolsCtrl.upgradeClient.ToolMod.MgoSession = &toolsCtrl.MgoClient
	toolsCtrl.upgradeClient.CourseMod.MgoSession = &toolsCtrl.MgoClient
}

// GetToolByName 通过工具名获取工具
func (toolsCtrl *ToolsController) GetToolByName() {
	name := toolsCtrl.Ctx.Input.Param(":name")
	tool, err := toolsCtrl.ToolMod.GetTool(name)
	if err != nil {
		toolsCtrl.abortWithError(m.ERR_TOOL_MESSAGE_QUERY_FAIL)
	}
	// 封装返回数据
	out := make(map[string]interface{})
	out["code"] = 0
	out["tool"] = tool
	toolsCtrl.jsonResult(out)
}

// GetToolsByWeight 获取工具列表通过权重
func (toolsCtrl *ToolsController) GetToolsByWeight() {
	paging, err := paramPaging(toolsCtrl.Ctx)
	if err != nil {
		toolsCtrl.abortWithError(m.ERR_REQUEST_PARAM)
	}
	tools, err := toolsCtrl.ToolMod.GetAllToolsByWeight(paging)
	if err != nil {
		toolsCtrl.abortWithError(m.ERR_TOOL_MESSAGE_QUERY_FAIL)
	}
	// 封装返回数据
	out := make(map[string]interface{})
	out["code"] = 0
	out["tools"] = tools
	toolsCtrl.jsonResult(out)
}

// GetTools 获取工具列表
func (toolsCtrl *ToolsController) GetTools() {
	tools, err := toolsCtrl.ToolMod.GetAllTools()
	if err != nil {
		toolsCtrl.abortWithError(m.ERR_TOOL_MESSAGE_QUERY_FAIL)
	}
	// 封装返回数据
	out := make(map[string]interface{})
	out["code"] = 0
	out["tools"] = tools
	toolsCtrl.jsonResult(out)
}

//GetToolsCategory 获取工具类型
func (toolsCtrl *ToolsController) GetToolsCategory() {
	categorys, err := toolsCtrl.ToolMod.GetToolsCategory()
	if err != nil {
		toolsCtrl.abortWithError(m.ERR_TOOL_MESSAGE_QUERY_FAIL)
	}
	// 封装返回数据
	out := make(map[string]interface{})
	out["code"] = 0
	out["categorys"] = categorys
	toolsCtrl.jsonResult(out)
}

//InstallTools 命令行安装工具
func (toolsCtrl *ToolsController) InstallTools() {

	if err := os.RemoveAll(path.Join(beego.AppPath, "asset", "tools")); err != nil {
		logs.Error("remove dir[%s] fail", path.Join(beego.AppPath, "asset", "tools"))
		return
	}
	url := beego.AppConfig.String("tool_mall_url")
	productKey, productSerial, err := daemon.GetProductInfo()
	if err != nil {
		logs.Error("GetProductInfo fail", err)
		return
	}
	if err := toolsCtrl.ToolMod.OrderInstallTools(url, productKey, productSerial); err != nil {
		logs.Error("Install tools fail", err)
		return
	}
	beego.Informational("Tools install complete!")
}

// CheckTools 获取需要升级的工具列表
func (toolsCtrl *ToolsController) CheckTools() {
	// // 获取token
	// token := toolsCtrl.checkToken()
	// // 检查是否拥有管理员的权限
	// toolsCtrl.needAdminPermission(token)
	tools, err := toolsCtrl.upgradeClient.CheckTools()
	//tools, err := m.GetToolsInfo()
	if err != nil {
		toolsCtrl.abortWithError(m.ERR_TOOL_MESSAGE_QUERY_FAIL)
	}
	// 封装返回数据
	out := make(map[string]interface{})
	out["code"] = 0
	out["tools"] = tools.Tools
	toolsCtrl.jsonResult(out)
}

//LaunchTools 升级工具
func (toolsCtrl *ToolsController) LaunchTools() {
	// // 获取token
	// token := toolsCtrl.checkToken()
	// // 检查是否拥有管理员的权限
	// toolsCtrl.needAdminPermission(token)
	ws, err := websocket.Upgrade(toolsCtrl.Ctx.ResponseWriter, toolsCtrl.Ctx.Request, nil, 1024, 1024)
	if err != nil {
		http.Error(toolsCtrl.Ctx.ResponseWriter, "Not a websocket handshake", 400)
	}
	tools, err := toolsCtrl.upgradeClient.CheckTools()
	if err != nil {
		toolsCtrl.abortWithError(m.ERR_TOOL_MESSAGE_QUERY_FAIL)
	}
	//如果有升级信息
	if tools.Newver {
		if err := toolsCtrl.ToolMod.DownloadTools(tools.Tools, ws); err != nil {
			logs.Error("DownloadTools is err:", err)
			toolsCtrl.abortWithError(m.ERR_TOOL_DOWNLOAD_FAIL)
		}
	}
	// 封装返回数据
	out := make(map[string]interface{})
	out["code"] = 0
	toolsCtrl.jsonResult(out)
}
