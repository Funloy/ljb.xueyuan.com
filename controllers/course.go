//@Descriptio: 课程管理控制器
//@Autho: yaohuarun
package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/astaxie/beego/logs"
	"github.com/gorilla/websocket"
	"gopkg.in/mgo.v2/bson"

	"github.com/astaxie/beego"
	m "maiyajia.com/models"
	"maiyajia.com/services/daemon"
)

type CouresController struct {
	BaseController
	CourseMod     m.CourseModels
	userMod       m.UserModels
	upgradeClient daemon.UpgradeModels
}

func (courseCtrl *CouresController) NestPrepare() {
	courseCtrl.CourseMod.MgoSession = &courseCtrl.MgoClient
	courseCtrl.userMod.MgoSession = courseCtrl.MgoClient
	courseCtrl.upgradeClient.MgoSession = &courseCtrl.MgoClient
	courseCtrl.upgradeClient.ToolMod.MgoSession = &courseCtrl.MgoClient
	courseCtrl.upgradeClient.CourseMod.MgoSession = &courseCtrl.MgoClient
}

//GetCoursesCategory 获取工具所属类型
func (courseCtrl *CouresController) GetCoursesCategory() {
	categorys, err := courseCtrl.CourseMod.GetCoursesCategory()
	if err != nil {
		logs.Error("GetCoursesCategory err:", err)
		courseCtrl.abortWithError(m.ERR_COURSE_MESSAGE_QUERY_FAIL)
	}
	// 封装返回数据
	out := make(map[string]interface{})
	out["code"] = 0
	out["categorys"] = categorys
	courseCtrl.jsonResult(out)
}

// GetLessions 获取课时列表并记录课程浏览量
func (courseCtrl *CouresController) GetLessions() {
	courseID := courseCtrl.Ctx.Input.Param(":courseId")
	if !bson.IsObjectIdHex(courseID) {
		courseCtrl.abortWithError(m.ERR_REQUEST_PARAM)
	}
	//记录课程浏览量
	if courseCtrl.CourseMod.RecordBrowseOfCourse(courseID) != nil {
		courseCtrl.abortWithError(m.ERR_COURSE_BROWSE_FAIL)
	}
	course, err := courseCtrl.CourseMod.GetAllLessions(bson.ObjectIdHex(courseID))
	if err != nil {
		courseCtrl.abortWithError(m.ERR_LESSIONS_MESSAGE_QUERY_FAIL)
	}
	out := make(map[string]interface{})
	out["code"] = 0
	out["course"] = course
	courseCtrl.jsonResult(out)
}

//InstallCourse 安装课程
func (courseCtrl *CouresController) InstallCourse() {

	url := beego.AppConfig.String("course_mall_url")
	logs.Info("url:", url)
	productKey, productSerial, err := daemon.GetProductInfo()
	if err != nil {
		logs.Error("GetProductInfo fail", err)
		return
	}
	if err = courseCtrl.CourseMod.OrderInstallCourses(url, productKey, productSerial); err != nil {
		logs.Error("Install Courses fail", err)
		return
	}
}

//CheckCourses 新增课程检测
func (courseCtrl *CouresController) CheckCourses() {
	// 获取token
	token := courseCtrl.checkToken()
	// 检查是否拥有管理员的权限
	courseCtrl.needAdminPermission(token)
	courses, err := courseCtrl.upgradeClient.CheckCourses()
	if err != nil {
		courseCtrl.abortWithError(m.ERR_COURSE_MESSAGE_QUERY_FAIL)
	}
	// 封装返回数据
	out := make(map[string]interface{})
	out["code"] = 0
	out["courses"] = courses.Courses
	courseCtrl.jsonResult(out)
}

//LaunchCourses 下载新增课程
func (courseCtrl *CouresController) LaunchCourses() {
	// // 获取token
	// token := courseCtrl.checkToken()
	// // 检查是否拥有管理员的权限
	// courseCtrl.needAdminPermission(token)
	ws, err := websocket.Upgrade(courseCtrl.Ctx.ResponseWriter, courseCtrl.Ctx.Request, nil, 1024, 1024)
	if err != nil {
		http.Error(courseCtrl.Ctx.ResponseWriter, "Not a websocket handshake", 400)
	}
	courses, err := courseCtrl.upgradeClient.CheckCourses()
	if err != nil {
		courseCtrl.abortWithError(m.ERR_COURSE_MESSAGE_QUERY_FAIL)
	}
	if courses.Newver {
		if err := courseCtrl.CourseMod.DownloadCourses(courses.Courses, ws); err != nil {
			logs.Error("DownloadCourses is err:", err)
			courseCtrl.abortWithError(m.ERR_COURSE_DOWNLOAD_FAIL)
		}
	}
	// 封装返回数据
	out := make(map[string]interface{})
	out["code"] = 0
	courseCtrl.jsonResult(out)
}

// AddCustomCourse 添加老师自定义课程
func (courseCtrl *CouresController) AddCustomCourse() {

	token := courseCtrl.checkToken()

	// 检查是否拥有创建班级的权限（角色为老师或管理员）
	courseCtrl.needAdminOrTeacherPermission(token)

	userID := token.UserID

	var courseParam m.CustomCourseParam

	// 凭证解析错误
	if err := json.Unmarshal(courseCtrl.Ctx.Input.RequestBody, &courseParam); err != nil {
		courseCtrl.abortWithError(m.ERR_REQUEST_PARAM)
	}

	_, err := courseCtrl.CourseMod.CustomCourseExit(userID, courseParam.CourseID[0])

	if err == nil {

		courseCtrl.abortWithError(m.ERR_CUSTOMECOURSE_EXIT)
	}

	customCourses := courseCtrl.CourseMod.NewCustomCourses(userID, courseParam.CourseID)

	for _, item := range customCourses {

		courseCtrl.CourseMod.RegisteredCustomCourses(item)
	}

	// 封装返回数据
	out := make(map[string]interface{})
	out["code"] = 0
	out["courses"] = customCourses
	courseCtrl.jsonResult(out)
}

// GetCustomCourse 查询课程中心列表
func (courseCtrl *CouresController) GetCustomCourse() {

	token := courseCtrl.checkToken()

	userID := token.UserID

	userName := token.Username

	if token.UserRole == m.ROLE_STUDENT {

		fmt.Println("student model:")

		var courses []m.CustomCourse

		c, err := courseCtrl.CourseMod.GetAllCustomCourses()

		if err != nil {

			courseCtrl.abortWithError(m.ERR_CUSTOMECOURSE_NULL)
		}

		for _, item := range c {

			// fmt.Println(item)

			for _, l := range item.Class {

				if courseCtrl.userMod.IsJoinedClass(userName, l.Code) == true {

					item.Class = []m.ClassPreview{}

					item.Class = append(item.Class, l)

					courses = append(courses, item)

					break
				}
			}
		}

		out := make(map[string]interface{})
		out["code"] = 0
		out["courses"] = courses
		courseCtrl.jsonResult(out)

	} else {

		c, err := courseCtrl.CourseMod.GetCustomCoursesByUserID(userID)

		if err != nil {

			courseCtrl.abortWithError(m.ERR_CUSTOMECOURSE_NULL)
		}

		out := make(map[string]interface{})
		out["code"] = 0
		out["courses"] = c
		courseCtrl.jsonResult(out)
	}

}

// RemoveCustomCourse 删除自定义课程
func (courseCtrl *CouresController) RemoveCustomCourse() {

	token := courseCtrl.checkToken()

	userID := token.UserID

	type courseBody struct {
		CourseID string `bson:"courseID" json:"courseID"`
	}

	var course courseBody

	// 凭证解析错误
	if err := json.Unmarshal(courseCtrl.Ctx.Input.RequestBody, &course); err != nil {
		courseCtrl.abortWithError(m.ERR_REQUEST_PARAM)
	}

	if err := courseCtrl.CourseMod.DeleteCustomCourses(userID, course.CourseID); err != nil {

		courseCtrl.abortWithError(m.ERR_CUSTOMECOURSE_DELETE)
	}

	out := make(map[string]interface{})
	out["code"] = 0
	courseCtrl.jsonResult(out)

}

// EditCustomCourse  编辑课程班级
func (courseCtrl *CouresController) EditCustomCourse() {

	token := courseCtrl.checkToken()

	userID := token.UserID

	var patchData m.PatchBody

	// 凭证解析错误
	if err := json.Unmarshal(courseCtrl.Ctx.Input.RequestBody, &patchData); err != nil {
		courseCtrl.abortWithError(m.ERR_REQUEST_PARAM)
	}

	fmt.Println(patchData)

	if err := courseCtrl.CourseMod.PatchCustomCourseClass(userID, patchData.CourseID, patchData.Class); err != nil {

		courseCtrl.abortWithError(m.ERR_CUSTOMECOURSE_DELETE)
	}

	out := make(map[string]interface{})
	out["code"] = 0
	out["back"] = patchData
	courseCtrl.jsonResult(out)

}

//GetClassCourse 根据班级查询所在班级课程
func (courseCtrl *CouresController) GetClassCourse() {
	var code string
	if courseCtrl.Ctx.Input.Bind(&code, "code") != nil {
		courseCtrl.abortWithError(m.ERR_REQUEST_PARAM)
	}
	result, err := courseCtrl.CourseMod.GetClassCourse(code)
	if err != nil {
		logs.Error("GetClassCourse(获取班级下的课程):", err)
		courseCtrl.abortWithError(m.ERR_CUSTOMECOURSE_DELETE)
	}
	out := make(map[string]interface{})
	out["code"] = 0
	out["result"] = result
	courseCtrl.jsonResult(out)
}

//GetCoursesByCategory 按课程分类查询课程
func (courseCtrl *CouresController) GetCoursesByCategory() {
	var courses interface{}
	// var sortCategory int
	var category string
	if courseCtrl.Ctx.Input.Bind(&category, "category") != nil {
		courseCtrl.abortWithError(m.ERR_REQUEST_PARAM)
	}
	// if courseCtrl.Ctx.Input.Bind(&sortCategory, "sortCategory") != nil {
	// 	courseCtrl.abortWithError(m.ERR_REQUEST_PARAM)
	// }
	paging, err := paramPaging(courseCtrl.Ctx)
	if err != nil {
		courseCtrl.abortWithError(m.ERR_REQUEST_PARAM)
	}
	total, _ := courseCtrl.CourseMod.QueryCoursesByCategoryCount(category)
	courses, err = courseCtrl.CourseMod.GetCoursesByCategory(paging, category)
	if err != nil {
		logs.Error("GetCoursesByCategory err:", err)
		courseCtrl.abortWithError(m.ERR_NO_WORK_EXISTS)
	}

	out := make(map[string]interface{})
	out["code"] = 0
	out["courses"] = courses
	out["total"] = total.Total
	courseCtrl.jsonResult(out)
}

// GetCourses 获取课程列表，不包括课程下的课时信息
func (courseCtrl *CouresController) GetCourses() {
	paging, err := paramPaging(courseCtrl.Ctx)
	if err != nil {
		courseCtrl.abortWithError(m.ERR_REQUEST_PARAM)
	}
	total, err := courseCtrl.CourseMod.QueryCoursesCount()
	if err != nil {
		logs.Error("QueryCoursesCount err :", err)
	}
	startTime := time.Now().Unix()
	courses, err := courseCtrl.CourseMod.GetAllCourses(paging)
	if err != nil {
		logs.Error("GetCourses err：", err)
		courseCtrl.abortWithError(m.ERR_COURSE_MESSAGE_QUERY_FAIL)
	}
	endTime := time.Now().Unix()
	logs.Info("GetAllCourses time:", endTime-startTime)
	// 封装返回数据
	out := make(map[string]interface{})
	out["code"] = 0
	out["courses"] = courses
	out["total"] = total
	courseCtrl.jsonResult(out)
}
