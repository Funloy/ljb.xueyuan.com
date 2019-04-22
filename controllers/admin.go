package controllers

import (
	"encoding/json"
	"net/http"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/gorilla/websocket"
	m "maiyajia.com/models"
	"maiyajia.com/services/daemon"
	"maiyajia.com/services/mongo"
)

type out struct {
	Code int `json:"code"`
	// Message string    `json:"message"`
	State *progress `json:"state"`
}
type progress struct {
	Stage    int `json:"stage"`
	Progress int `json:"progress"`
}

// AdminController 控制器
type AdminController struct {
	BaseController
	userMod     m.UserModels
	toolMod     m.ToolModels
	courseMod   m.CourseModels
	upgradeMod  daemon.UpgradeModels
	livenessMod daemon.LivenessModel
	accountMod  daemon.AccountModels
}

// NestPrepare 初始化函数
// 把控制器的MgoClient赋值到模型的数据库操作客户端
func (adminCtl *AdminController) NestPrepare() {
	mongo.Client = &adminCtl.MgoClient
	adminCtl.userMod.MgoSession = adminCtl.MgoClient
	adminCtl.toolMod.MgoSession = &adminCtl.MgoClient
	adminCtl.courseMod.MgoSession = &adminCtl.MgoClient
	adminCtl.upgradeMod.CourseMod.MgoSession = &adminCtl.MgoClient
	adminCtl.upgradeMod.ToolMod.MgoSession = &adminCtl.MgoClient
}

// GetSystemInfo 获取系统信息
func (adminCtl *AdminController) GetSystemInfo() {

	// 封装返回数据
	out := make(map[string]interface{})
	out["code"] = 0
	out["memory"] = daemon.QueryMemUsage()
	out["disk"] = daemon.QueryDiskUsage()
	adminCtl.jsonResult(out)
}

// GetSystemAccount 获取产品信息
func (adminCtl *AdminController) GetSystemAccount() {

	// 封装返回数据
	out := make(map[string]interface{})
	out["code"] = 0
	out["account"] = daemon.FetchSystemAccount()
	adminCtl.jsonResult(out)
}

// CheckSystemUpgrade 升级检查
func (adminCtl *AdminController) CheckSystemUpgrade() {
	// 获取token
	token := adminCtl.checkToken()
	if token.UserRole != m.ROLE_ADMIN {
		adminCtl.abortWithError(m.ERR_ADMIN_UNPERMIT)
	}
	var procourses []m.Course
	_, upgrade, err := daemon.UpgradeCheck()
	if err != nil {
		adminCtl.abortWithError(m.ERR_UPGRADE_CHECK_FAIL)
	}
	courses, err := adminCtl.upgradeMod.CheckCourses()
	if err != nil {
		adminCtl.abortWithError(m.ERR_COURSE_MESSAGE_QUERY_FAIL)
	}
	for _, cou := range courses.Courses {
		if cou.Purchased == false {
			continue
		}
		procourses = append(procourses, cou)
	}
	tools, err := adminCtl.upgradeMod.CheckTools()
	if err != nil {
		adminCtl.abortWithError(m.ERR_TOOL_MESSAGE_QUERY_FAIL)
	}
	// // 延迟处理的函数
	// defer adminCtl.Recover(m.RUNTIME_ERROR)

	if upgrade == nil {
		// 封装返回数据
		out := make(map[string]interface{})
		out["code"] = 0
		out["system"] = upgrade
		out["courses"] = procourses
		out["tools"] = tools.Tools
		adminCtl.jsonResult(out)
	}
	// 封装返回数据
	out := make(map[string]interface{})
	out["code"] = 0
	out["system"] = upgrade
	out["courses"] = procourses
	out["tools"] = tools.Tools
	adminCtl.jsonResult(out)
}

//LaunchSystemUpgrade 平台升级
func (adminCtl *AdminController) LaunchSystemUpgrade() {
	ws, err := websocket.Upgrade(adminCtl.Ctx.ResponseWriter, adminCtl.Ctx.Request, nil, 1024, 1024)
	if err != nil {
		http.Error(adminCtl.Ctx.ResponseWriter, "Not a websocket handshake", 400)
	}
	err = daemon.LaunchUpgrade(ws)
	if err != nil {
		beego.Error(err)
		adminCtl.abortWithError(m.ERR_UPGRADE_FAIL)
	}
}

//RebootSystemUpgrade 重启升级平台系统
func (adminCtl *AdminController) RebootSystemUpgrade() {
	if err := daemon.UpdateUpgradeInfo(); err != nil {
		adminCtl.abortWithError(m.ERR_UPGRADE_REBOOT)
	}
	if err := daemon.SysReboot(); err != nil {
		adminCtl.abortWithError(m.ERR_UPGRADE_REBOOT)
	}
}

// GetApplicationlist 资格申请列表展示（管理员权限）
func (adminCtl *AdminController) GetApplicationlist() {
	// 获取token
	token := adminCtl.checkToken()
	if token.UserRole != m.ROLE_ADMIN {
		adminCtl.abortWithError(m.ERR_ADMIN_UNPERMIT)
	}
	qualifylist, err := adminCtl.userMod.QueryApplicationList()
	if err != nil {
		adminCtl.abortWithError(m.ERR_ADMIN_APPLICATION_QUERY_FAIL)
	}

	// 封装返回数据
	out := make(map[string]interface{})
	out["code"] = 0
	out["qualifylist"] = qualifylist
	adminCtl.jsonResult(out)
}

//ReviewApplication 审核申请（管理员权限）
func (adminCtl *AdminController) ReviewApplication() {
	// 获取token
	token := adminCtl.checkToken()
	if token.UserRole != m.ROLE_ADMIN {
		adminCtl.abortWithError(m.ERR_ADMIN_UNPERMIT)
	}
	var review m.Review
	if err := json.Unmarshal(adminCtl.Ctx.Input.RequestBody, &review); err != nil {
		adminCtl.abortWithError(m.ERR_REQUEST_PARAM)
	}
	status := review.Status
	userid := review.UserID
	if status == 0 {
		//拒绝申请
		if err := adminCtl.userMod.UpdateAplStatus(status, userid); err != nil {
			adminCtl.abortWithError(m.ERR_REVIEW_APPLICATION_FAIL)
		}
		// 封装返回数据
		out := make(map[string]interface{})
		out["code"] = 0
		adminCtl.jsonResult(out)
	} else if status == 1 {
		//同意申请
		if err := adminCtl.userMod.UpdateUserRole(userid); err != nil {
			adminCtl.abortWithError(m.ERR_REVIEW_APPLICATION_FAIL)
		}
		if err := adminCtl.userMod.UpdateAplStatus(status, userid); err != nil {
			adminCtl.abortWithError(m.ERR_REVIEW_APPLICATION_FAIL)
		}

		// result, err := m.UpdateStatusRole(userid)
		// if err != nil {
		// 	logs.Info("UpdateStatusRole err:", err)
		// }
		// 封装返回数据
		out := make(map[string]interface{})
		out["code"] = 0
		adminCtl.jsonResult(out)
	}
}

//DownloadData 下载课程、工具（管理员权限）
func (adminCtl *AdminController) DownloadData() {
	var out out
	token := adminCtl.checkToken()
	if token.UserRole != m.ROLE_ADMIN {
		adminCtl.abortWithError(m.ERR_ADMIN_UNPERMIT)
	}
	ws, err := websocket.Upgrade(adminCtl.Ctx.ResponseWriter, adminCtl.Ctx.Request, nil, 1024, 1024)
	if err != nil {
		http.Error(adminCtl.Ctx.ResponseWriter, "Not a websocket handshake", 400)
	}
	//清除课程，工具数据
	if err := os.RemoveAll(path.Join(beego.AppPath, "asset", "tools")); err != nil {
		logs.Error("remove dir[%s] fail", path.Join(beego.AppPath, "asset", "tools"))
		return
	}
	if err := os.RemoveAll(path.Join(beego.AppPath, "asset", "courses")); err != nil {
		logs.Error("remove dir[%s] fail", path.Join(beego.AppPath, "asset", "courses"))
		return
	}
	productKey, productSerial, err := daemon.GetProductInfo()
	if err != nil {
		logs.Error("GetProductInfo fail", err)
		return
	}

	//下载工具
	logs.Info("begin install tools")
	url := beego.AppConfig.String("tool_mall_url")
	if err := adminCtl.toolMod.InstallTools(url, productKey, productSerial, ws); err != nil {
		logs.Error("Install tools fail", err)
		send_msg, _ := json.Marshal(daemon.WsAbortWithError(m.ERR_TOOL_DOWNLOAD_FAIL))
		ws.WriteMessage(websocket.TextMessage, send_msg)
		return
	}
	out.Code = 0
	out.State = &progress{
		Stage:    1,
		Progress: 50,
	}
	send_msg, err := json.Marshal(out)
	if err != nil {
		logs.Error("error:", err)
	}
	ws.WriteMessage(websocket.TextMessage, send_msg)
	logs.Info("install tools complete")
	//下载课程
	logs.Info("begin install courses")
	url = beego.AppConfig.String("course_mall_url")
	if err = adminCtl.courseMod.InstallCourses(url, productKey, productSerial, ws); err != nil {
		logs.Error("Install Courses fail", err)
		send_msg, _ := json.Marshal(daemon.WsAbortWithError(m.ERR_COURSE_DOWNLOAD_FAIL))
		ws.WriteMessage(websocket.TextMessage, send_msg)
		return
	}
	out.State = &progress{
		Stage:    2,
		Progress: 100,
	}
	send_msg, err = json.Marshal(out)
	if err != nil {
		logs.Error("error:", err)
	}
	ws.WriteMessage(websocket.TextMessage, send_msg)
	beego.Informational("Courses install complete!")
}

//GetLivenessCount 用户活跃度统计
func (adminCtl *AdminController) GetLivenessCount() {

	startyear, _ := strconv.Atoi(adminCtl.Ctx.Input.Param(":startyear"))
	startmonth, _ := strconv.Atoi(adminCtl.Ctx.Input.Param(":startmonth"))
	endyear, _ := strconv.Atoi(adminCtl.Ctx.Input.Param(":endyear"))
	endmonth, _ := strconv.Atoi(adminCtl.Ctx.Input.Param(":endmonth"))
	startTime := time.Date(startyear, time.Month(startmonth), 1, 0, 0, 0, 0, time.UTC)
	deadlineTime := time.Date(endyear, time.Month(endmonth+1), 1, 0, 0, 0, 0, time.UTC)
	go func() {
		adminCtl.livenessMod.UserMod.MgoSession = adminCtl.MgoClient
		adminCtl.livenessMod.UserMod.MgoSession.StartSession()
		defer adminCtl.livenessMod.UserMod.MgoSession.CloseSession()
		adminCtl.livenessMod.UpdateUserLiveness(endyear, endmonth)
	}()
	liveness, err := adminCtl.userMod.GetUserLive(startTime, deadlineTime)
	if err != nil {
		adminCtl.abortWithError(m.ERR_GETUSERLIVE_FAIL)
	}
	userscount, err := adminCtl.userMod.GetUsersCount()
	if err != nil {
		logs.Error("GetUsersCount err:", err)
		adminCtl.abortWithError(m.ERR_GETUSERSCOUNT_FAIL)
	}
	// 封装返回数据
	out := make(map[string]interface{})
	out["code"] = 0
	out["usersCount"] = userscount
	out["liveness"] = liveness
	adminCtl.jsonResult(out)
}

//InsertLiveness 初次用户活跃度统计记录
func (adminCtl *AdminController) InsertLiveness() {

	// startyear, _ := strconv.Atoi(adminCtl.Ctx.Input.Param(":startyear"))
	// startmonth, _ := strconv.Atoi(adminCtl.Ctx.Input.Param(":startmonth"))
	startyear := 2018
	startmonth := 4
	endtime := time.Now()
	startTime := time.Date(startyear, time.Month(startmonth), 1, 0, 0, 0, 0, time.UTC)
	months := (endtime.Year()-startyear)*12 + int(endtime.Month()) - startmonth + 1
	if err := adminCtl.userMod.DellivenessData(); err != nil {

	}
	var liveness []m.ActiveCount
	for index := 0; index < months; index++ {
		monStart := startTime.AddDate(0, index, 0)
		monEnd := startTime.AddDate(0, index+1, 0)
		total, err := adminCtl.userMod.GetLivenessCount(monStart.Unix(), monEnd.Unix())
		if err != nil {
			adminCtl.abortWithError(m.ERR_REVIEW_APPLICATION_FAIL)
		}
		total.Time = monStart
		liveness = append(liveness, total)
	}
	if err := adminCtl.userMod.InsertUserLive(liveness); err != nil {
		adminCtl.abortWithError(m.ERR_INSERTUSERLIVE_FAIL)
	}
	// 封装返回数据
	out := make(map[string]interface{})
	out["code"] = 0
	adminCtl.jsonResult(out)
}

/*********************************************************************************************/
/*********************************** 以下为本控制器的内部函数 *********************************/
/*********************************** *********************************************************/
