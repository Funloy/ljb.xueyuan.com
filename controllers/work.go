package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"path"
	"strings"

	"gopkg.in/mgo.v2/bson"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"

	m "maiyajia.com/models"
	"maiyajia.com/services/daemon"
	d "maiyajia.com/services/daemon"
)

// WorksController 作品控制器
type WorksController struct {
	BaseController
	workMod m.WorkModels
	userMod m.UserModels
	toolMod m.ToolModels
}

// NestPrepare 数据库客户端
func (workCtrl *WorksController) NestPrepare() {
	workCtrl.workMod.MgoSession = &workCtrl.MgoClient
	workCtrl.userMod.MgoSession = workCtrl.MgoClient
	workCtrl.toolMod.MgoSession = &workCtrl.MgoClient
}

// GetList 获取作品列表
func (workCtrl *WorksController) GetList() {
	workCtrl.checkToken()
	var err error

	// 获取请求参数中的分页数据

	paging, err := paramPaging(workCtrl.Ctx)

	if err != nil {

		workCtrl.abortWithError(m.ERR_REQUEST_PARAM)
	}
	userID := workCtrl.Ctx.Input.Param(":userid")
	if !bson.IsObjectIdHex(userID) {
		workCtrl.abortWithError(m.ERR_REQUEST_PARAM)
	}

	total, err := workCtrl.workMod.QueryWorksCount(userID, "works")
	if err != nil {
		logs.Error("QueryWorksCount err:", err)
		workCtrl.abortWithError(m.ERR_COUNT_FAIL)
	}
	tool := workCtrl.GetString("tool")

	var works []m.WorkListBody

	tool = strings.ToLower(tool)

	works, err = workCtrl.workMod.GetWorksByID(userID, tool, paging)

	if err != nil {

		workCtrl.abortWithError(m.ERR_NO_WORK_EXISTS)
	}

	out := make(map[string]interface{})
	out["code"] = 0
	out["works"] = works
	out["total"] = total.Total
	workCtrl.jsonResult(out)

}

//PostDescription 保存作品
func (workCtrl *WorksController) PostDescription() {
	workCtrl.checkToken()
	var workContent m.WorkBody

	beego.Info("begin desc")
	// // 验证提交的课程信息: 验证课程信息有效性
	if err := json.Unmarshal(workCtrl.Ctx.Input.RequestBody, &workContent); err != nil {
		beego.Error("params error:", err)
		workCtrl.abortWithError(m.ERR_REQUEST_PARAM)
	}

	if workContent.ID.Hex() == "" {

		workContent.ID = m.NewIDBson()
	}
	toolOBJ, e := workCtrl.toolMod.GetTool(workContent.Tool)
	if e == nil {
		workContent.ToolURL = toolOBJ.Relpath
	}
	workContent.Category = toolOBJ.Category
	var relpath string
	if workContent.Types == "stl" {
		name := workContent.ID.Hex() + ".stl"
		relpath = path.Join("asset", "works", name)
	}
	newWorkContent := m.NewWork(workContent.ID, workContent.UserID, workContent.ContentID.Hex(), workContent.Name, workContent.Tool, workContent.Types, relpath, workContent.Picture, workContent.Description, workContent.Data, workContent.ToolURL, workContent.Category, workContent.Public)

	if err := workCtrl.workMod.RegisteredWork(newWorkContent); err != nil {

		workCtrl.abortWithError(m.ERR_ADD_WORK_FAIL)

	}

	beego.Debug("end desc")

	out := make(map[string]interface{})
	out["code"] = 0
	workCtrl.jsonResult(out)
}

// PostBinaryData 提交二进制文件（stl,sgl）
func (workCtrl *WorksController) PostBinaryData() {
	var id, name string
	workCtrl.checkToken()
	if workCtrl.Ctx.Input.Bind(&id, "id") != nil {
		workCtrl.abortWithError(m.ERR_REQUEST_PARAM)
	}
	if id != "" {
		name = id + ".sgl"
	} else {
		id = m.NewID()
		name = id + ".stl"
	}
	beego.Info("begin PostBinaryData")
	content := workCtrl.Ctx.Input.RequestBody
	workpath := path.Join(beego.AppPath, "asset", "works", name)

	dirPath := path.Join(beego.AppPath, "asset", "works")

	_, e := os.Stat(dirPath)

	if e != nil {

		err := os.Mkdir(dirPath, os.ModePerm)
		if err != nil {
			fmt.Printf("mkdir failed![%v]\n", err)

			workCtrl.abortWithError(m.ERR_CREATE_FILE_FAIL)

		} else {
			fmt.Printf("mkdir success!\n")
		}
	}

	fp, err := os.Create(workpath)

	if err != nil {

		beego.Debug("fail to create the file")
		workCtrl.abortWithError(m.ERR_CREATE_FILE_FAIL)
	}

	defer fp.Close()
	fp.Write(content)
	out := make(map[string]interface{})

	var responData m.Data
	responData.ID = id
	out["code"] = 0
	out["data"] = responData
	workCtrl.jsonResult(out)
	beego.Info("end PostBinaryData")
}

//PutDescription 更新作品
func (workCtrl *WorksController) PutDescription() {
	workCtrl.checkToken()
	var workContent m.WorkBody

	beego.Debug("begin desc")

	// // 验证提交的课程信息: 验证课程信息有效性
	if err := json.Unmarshal(workCtrl.Ctx.Input.RequestBody, &workContent); err != nil {
		beego.Debug("params error:")

		workCtrl.abortWithError(m.ERR_REQUEST_PARAM)

	}

	if err := workCtrl.workMod.PatchWork(workContent.ID.Hex(), workContent.UserID, workContent.Name, workContent.Tool, workContent.Types, workContent.Picture, workContent.Description, workContent.Data); err != nil {

		workCtrl.abortWithError(m.ERR_ADD_WORK_FAIL)

	}

	beego.Debug("end desc")

	out := make(map[string]interface{})
	out["code"] = 0
	workCtrl.jsonResult(out)
}

// PutBinaryData 更新二进制文件
func (workCtrl *WorksController) PutBinaryData() {
	workCtrl.checkToken()
	var name string
	id := workCtrl.GetString("id")
	suffix := workCtrl.GetString("suffix")
	if suffix == "sgl" {
		name = id + ".sgl"
	} else if suffix == "stl" {
		name = id + ".stl"
	}
	beego.Info("begin PostBinaryData")
	content := workCtrl.Ctx.Input.RequestBody
	path := path.Join(beego.AppPath, "asset", "works", name)
	beego.Debug(path)
	fp, _ := os.Create(path)
	defer fp.Close()
	fp.Write(content)
	out := make(map[string]interface{})

	var responData m.Data
	responData.ID = id
	out["code"] = 0
	out["data"] = responData
	workCtrl.jsonResult(out)
}

// DeleteWork 删除作品
func (workCtrl *WorksController) DeleteWork() {
	workCtrl.checkToken()
	id := workCtrl.GetString("id")
	suffix := workCtrl.GetString("suffix")

	//检测作品ID是否存在
	work, err := workCtrl.workMod.FindWorkByID(bson.ObjectIdHex(id))

	if err != nil {
		logs.Info("查找失败")
		workCtrl.abortWithError(m.ERR_NO_WORK_EXISTS)
	}
	if workCtrl.workMod.FindWorkFavor(id) != nil {
		workCtrl.workMod.DeleteFavorWorkByID(id)
	}
	err = workCtrl.workMod.DeleteWorkByID(id)
	if err != nil {
		logs.Info("删除失败")
		workCtrl.abortWithError(m.ERR_DELETE_WORK_FAIL)
	}
	if suffix == "sgl" {
		name := id + ".sgl"
		err = os.Remove(path.Join(beego.AppPath, "asset", "works", name))
		if err != nil {
			workCtrl.abortWithError(m.ERR_SGL_NO_file_EXISTS)
		}
	}
	if work.Relpath != "" {
		err = os.Remove(work.Relpath)
		if err != nil {
			workCtrl.abortWithError(m.ERR_NO_file_EXISTS)
		}
	}
	//封装返回数据
	out := make(map[string]interface{})
	out["code"] = 0
	// 返回结果
	workCtrl.jsonResult(out)

}

// GetWork 获取作品二进制文件
func (workCtrl *WorksController) GetWork() {
	workCtrl.checkToken()
	var filename string
	id := workCtrl.GetString("id")
	suffix := workCtrl.GetString("suffix")
	if id == "" {
		workCtrl.abortWithError(m.ERR_NO_WORK_EXISTS)
	}
	if suffix == "sgl" {
		filename = id + ".sgl"
	} else if suffix == "stl" {
		filename = id + ".stl"
	}

	fullPath := path.Join(beego.AppPath, "asset", "works", filename)

	_, err := os.Stat(fullPath)

	if err != nil {

		workCtrl.abortWithError(m.ERR_NO_WORK_EXISTS)

	}

	workCtrl.Ctx.Output.Download(path.Join(beego.AppPath, "asset", "works", filename))

}

// GetDesc 获取描述
func (workCtrl *WorksController) GetDesc() {
	workCtrl.checkToken()
	id := workCtrl.GetString("id")
	if id == "" {
		workCtrl.abortWithError(m.ERR_NO_WORK_EXISTS)
	}

	beego.Debug(id)

	workDesc, err := workCtrl.workMod.GetDesc(id)

	if err != nil {
		workCtrl.abortWithError(m.ERR_NO_WORK_EXISTS)
	}

	out := make(map[string]interface{})
	out["code"] = 0
	out["works"] = workDesc
	workCtrl.jsonResult(out)

}

// CreateID  创建作品ID
func (workCtrl *WorksController) CreateID() {
	workCtrl.checkToken()
	out := make(map[string]interface{})
	out["code"] = 0
	out["id"] = m.NewID()
	// 返回结果
	workCtrl.jsonResult(out)
}

//ShareWork 分享作品
func (workCtrl *WorksController) ShareWork() {
	// 获取token
	token := workCtrl.checkToken()
	var sharework m.ShareWork
	if err := json.Unmarshal(workCtrl.Ctx.Input.RequestBody, &sharework); err != nil {
		workCtrl.abortWithError(m.ERR_REQUEST_PARAM)

	}
	if err := workCtrl.workMod.PublicWork(sharework.ID, token.UserID, sharework.Public, sharework.Edit); err != nil {
		workCtrl.abortWithError(m.ERR_SHARE_WORK_FAIL)

	}

	out := make(map[string]interface{})
	out["code"] = 0
	workCtrl.jsonResult(out)
}

// GetShareListByCategory 获取分享作品列表按分类展示
func (workCtrl *WorksController) GetShareListByCategory() {
	var works interface{}
	var sortCategory int
	var category string
	if workCtrl.Ctx.Input.Bind(&category, "category") != nil {
		workCtrl.abortWithError(m.ERR_REQUEST_PARAM)
	}
	if workCtrl.Ctx.Input.Bind(&sortCategory, "sortCategory") != nil {
		workCtrl.abortWithError(m.ERR_REQUEST_PARAM)
	}
	paging, err := paramPaging(workCtrl.Ctx)
	if err != nil {
		workCtrl.abortWithError(m.ERR_REQUEST_PARAM)
	}
	total, _ := workCtrl.workMod.QueryWorksByCategoryCount(category)
	works, err = workCtrl.workMod.GetWorksByCategory(paging, sortCategory, category)
	if err != nil {
		workCtrl.abortWithError(m.ERR_NO_WORK_EXISTS)
	}

	out := make(map[string]interface{})
	out["code"] = 0
	out["works"] = works
	out["total"] = total.Total
	workCtrl.jsonResult(out)
}

//GetLaudAndFavorCou 获取个人作品点赞和被收藏总数
func (workCtrl *WorksController) GetLaudAndFavorCou() {
	var userID string
	var favorCount, laudCount = 0, 0
	if workCtrl.Ctx.Input.Bind(&userID, "userID") != nil {
		workCtrl.abortWithError(m.ERR_REQUEST_PARAM)
	}
	perShareCous, err := workCtrl.workMod.GetPersonShareWorksCou(userID)
	if err != nil {
		workCtrl.abortWithError(m.ERR_NO_WORK_EXISTS)
	}
	for _, perShareCou := range perShareCous {
		favorCount += perShareCou.Favor
		laudCount += perShareCou.Laud
	}

	out := make(map[string]interface{})
	out["code"] = 0
	out["favorCount"] = favorCount
	out["laudCount"] = laudCount
	workCtrl.jsonResult(out)
}

//GetPersonalShareList 获取指定个人分享的作品列表
func (workCtrl *WorksController) GetPersonalShareList() {
	var works []m.WorkListBody
	var userID string
	var user *m.User
	// toolsMap := make(map[string]string)
	if workCtrl.Ctx.Input.Bind(&userID, "userID") != nil {
		workCtrl.abortWithError(m.ERR_REQUEST_PARAM)
	}
	paging, err := paramPaging(workCtrl.Ctx)
	if err != nil {
		workCtrl.abortWithError(m.ERR_REQUEST_PARAM)
	}
	total, _ := workCtrl.workMod.QueryPersonShareWorksCount(userID)
	works, err = workCtrl.workMod.GetPersonShareWorks(paging, userID)
	if err != nil {
		workCtrl.abortWithError(m.ERR_NO_WORK_EXISTS)
	}
	if len(works) > 0 {

		user, _ = workCtrl.userMod.FetchGivenUserInfo(works[0].UserID, "realname")
		logs.Info("user:", user.Realname)
	}
	for i := 0; i < len(works); i++ {
		works[i].RealName = user.Realname
	}
	out := make(map[string]interface{})
	out["code"] = 0
	out["works"] = works
	out["total"] = total.Total
	workCtrl.jsonResult(out)
}

// GetShareList 获取分享作品列表按浏览量、时间展示
func (workCtrl *WorksController) GetShareList() {
	var works interface{}
	var sortCategory int
	if workCtrl.Ctx.Input.Bind(&sortCategory, "sortCategory") != nil {
		workCtrl.abortWithError(m.ERR_REQUEST_PARAM)
	}
	paging, err := paramPaging(workCtrl.Ctx)
	if err != nil {
		workCtrl.abortWithError(m.ERR_REQUEST_PARAM)
	}
	total, _ := workCtrl.workMod.QueryShareWorksCount()
	works, err = workCtrl.workMod.GetShareWorks(paging, sortCategory)
	if err != nil {
		workCtrl.abortWithError(m.ERR_NO_WORK_EXISTS)
	}

	out := make(map[string]interface{})
	out["code"] = 0
	out["works"] = works
	out["total"] = total.Total
	workCtrl.jsonResult(out)
}

// GetPreview 获取描述
func (workCtrl *WorksController) GetPreview() {

	id := workCtrl.GetString("id")
	if !bson.IsObjectIdHex(id) {
		logs.Error("workid is not ObjectId")
		workCtrl.abortWithError(m.ERR_REQUEST_PARAM)
	}
	workDesc, err := workCtrl.workMod.GetPreview(id)
	if err != nil {
		logs.Error("work is not exists")
		workCtrl.abortWithError(m.ERR_NO_WORK_EXISTS)
	}
	username, ok := workDesc["username"].([]interface{})
	avatar, ok := workDesc["avatar"].([]interface{})
	if ok {
		workDesc["username"] = username[0]
		workDesc["avatar"] = avatar[0]
	}
	out := make(map[string]interface{})
	out["code"] = 0
	out["works"] = workDesc
	workCtrl.jsonResult(out)

}

//GiveLaud 点赞作品
func (workCtrl *WorksController) GiveLaud() {
	var recordLaud m.RecordFavorOrLaud
	if err := json.Unmarshal(workCtrl.Ctx.Input.RequestBody, &recordLaud); err != nil {
		workCtrl.abortWithError(m.ERR_REQUEST_PARAM)
	}
	if recordLaud.UserID.Hex() == "" {
		recordLaud.UserID = bson.NewObjectId()
	}
	workID := recordLaud.WorkID
	userID := recordLaud.UserID
	if _, err := workCtrl.workMod.FindIsLaud(workID, userID); err == nil {
		workCtrl.abortWithError(m.ERR_LAUD_RECORD_QUERY_FAIL)
	}
	recordLaud.ID = bson.NewObjectId()
	if err := workCtrl.workMod.InsertLaud(recordLaud); err != nil {
		logs.Info("记录点赞失败")
		workCtrl.abortWithError(m.ERR_LAUD_WORK_FAIL)
	}
	if err := workCtrl.workMod.GiveLaud(workID); err != nil {
		logs.Info("点赞失败")
		workCtrl.abortWithError(m.ERR_LAUD_WORK_FAIL)
	}
	out := make(map[string]interface{})
	out["code"] = 0
	out["userid"] = userID
	workCtrl.jsonResult(out)
}

//FavoriteWork 收藏作品
func (workCtrl *WorksController) FavoriteWork() {
	token := workCtrl.checkToken()
	var recordFavor m.RecordFavorOrLaud
	if err := json.Unmarshal(workCtrl.Ctx.Input.RequestBody, &recordFavor); err != nil {
		workCtrl.abortWithError(m.ERR_REQUEST_PARAM)
	}
	workID := recordFavor.WorkID
	userID := token.UserID
	work, err := workCtrl.workMod.FindWorkByID(workID)
	//本人无法收藏
	if work.UserID.Hex() == token.UserID {
		workCtrl.abortWithError(m.ERR_SELF_UNCOLLECTION_FAIL)
	}
	if err != nil {
		workCtrl.abortWithError(m.ERR_NO_WORK_EXISTS)
	}
	//记录收藏信息
	if _, err := workCtrl.workMod.FindIsFacvor(workID, bson.ObjectIdHex(userID)); err == nil {
		workCtrl.abortWithError(m.ERR_QUERY_FAVOR_FAIL)
	}
	recordFavor.ID = bson.NewObjectId()
	recordFavor.UserID = bson.ObjectIdHex(userID)
	logs.Info("recordFavor:", recordFavor)
	if err := workCtrl.workMod.InsertFavor(recordFavor); err != nil {
		workCtrl.abortWithError(m.ERR_COLLECTION_WORK_FAIL)
	}
	if err := workCtrl.workMod.FavorCount(workID); err != nil {
		workCtrl.abortWithError(m.ERR_COLLECTION_WORK_FAIL)
	}
	out := make(map[string]interface{})
	out["code"] = 0
	workCtrl.jsonResult(out)
}

//GetMyFavoriteWorks 获取个人收藏列表
func (workCtrl *WorksController) GetMyFavoriteWorks() {
	//token := workCtrl.checkToken()
	results := []bson.M{}
	userID := workCtrl.GetString("userID")
	paging, err := paramPaging(workCtrl.Ctx)
	if err != nil {
		workCtrl.abortWithError(m.ERR_REQUEST_PARAM)
	}
	total, _ := workCtrl.workMod.QueryWorksCount(userID, "favor")
	favors := workCtrl.workMod.FindUserWorkFavor(userID, paging)
	if len(favors) > 0 {
		for _, favor := range favors {
			workid := favor.WorkID
			result, err := workCtrl.workMod.GetUserFavoriteWork(workid)
			if err != nil {
				logs.Info("GetUserFavoriteWork err:", err)
				workCtrl.abortWithError(m.ERR_FAVORITE_WORK_LIST)
			}
			results = append(results, result)
		}
	}
	out := make(map[string]interface{})
	out["code"] = 0
	out["favorites"] = results
	out["total"] = total.Total
	workCtrl.jsonResult(out)

}

//CopyWork 复制收藏作品用于编辑
func (workCtrl *WorksController) CopyWork() {
	logs.Info("复制收藏")
	//复制作品不需要stl的工具名
	toolName := []string{"mpython", "imakeduino"}
	id := workCtrl.GetString("workid")
	name := workCtrl.GetString("name")
	newid := m.NewIDBson()
	token := workCtrl.checkToken()
	work, err := workCtrl.workMod.FindWorkByID(bson.ObjectIdHex(id))
	if err != nil {
		workCtrl.abortWithError(m.ERR_NO_WORK_EXISTS)
	}
	if err := workCtrl.workMod.CopyWork(work, token.UserID, id, name, newid, toolName); err != nil {
		workCtrl.abortWithError(m.ERR_COLLECTION_WORK_FAIL)
	}
	if work.Types == "stl" {
		name := newid.Hex() + ".stl"
		dstName := path.Join(beego.AppPath, "asset", "works", name)
		srcName := path.Join(beego.AppPath, work.Relpath)
		daemon.CopyFile(dstName, srcName)
	}
	if work.Tool == "imakeart" {
		name := newid.Hex() + ".sgl"
		srcname := id + ".sgl"
		dstName := path.Join(beego.AppPath, "asset", "works", name)
		srcName := path.Join(beego.AppPath, "asset", "works", srcname)
		daemon.CopyFile(dstName, srcName)
	}
	if work.Tool == "scratch" {
		name := newid.Hex() + ".sb3"
		srcname := id + ".sb3"
		dstName := path.Join(beego.AppPath, "asset", "works", name)
		srcName := path.Join(beego.AppPath, "asset", "works", srcname)
		daemon.CopyFile(dstName, srcName)
	}

	out := make(map[string]interface{})
	out["code"] = 0
	out["id"] = newid
	workCtrl.jsonResult(out)
}

//DeleteFavoriteWork 删除个人收藏的作品
func (workCtrl *WorksController) DeleteFavoriteWork() {
	workid := workCtrl.GetString("workid")
	userid := workCtrl.checkToken().UserID
	logs.Info(workid)
	if err := workCtrl.workMod.DeleteFavoriteWork(workid, userid); err != nil {
		logs.Error("DeleteFavoriteWork err:", err)
		workCtrl.abortWithError(m.ERR_DELETE_FAVORITE_WORK)
	}
	out := make(map[string]interface{})
	out["code"] = 0
	workCtrl.jsonResult(out)
}

//RecordBrowse 记录作品浏览量
func (workCtrl *WorksController) RecordBrowse() {
	workid := workCtrl.GetString("workid")
	if workCtrl.workMod.RecordBrowse(workid) != nil {
		workCtrl.abortWithError(m.ERR_COURSE_BROWSE_FAIL)
	}
	out := make(map[string]interface{})
	out["code"] = 0
	workCtrl.jsonResult(out)
}

// SaveProject 保存scratch项目
func (workCtrl *WorksController) SaveProject() {

	token := workCtrl.checkToken()

	name := workCtrl.GetString("name")
	description := workCtrl.GetString("description")
	picture := workCtrl.GetString("picture")
	types := workCtrl.GetString("types")
	tool := workCtrl.GetString("tool")

	if !(name != "" && description != "" && picture != "" && types != "" && tool != "") {

		workCtrl.abortWithError(m.ERR_REQUEST_PARAM)
	}

	workDic := "./asset/works"

	b, _ := d.PathExists(workDic)

	if !b {

		err := os.Mkdir(workDic, os.ModePerm)

		if err != nil {

			workCtrl.abortWithError(m.ERR_DIRECTORY_CREATE)
		}
	}

	toolURL := ""

	toolOBJ, e := workCtrl.toolMod.GetTool(tool)

	if e == nil {

		toolURL = toolOBJ.Relpath

	} else {

		workCtrl.abortWithError(m.ERR_TOOL_REALPATH)

	}

	newWorkContent := workCtrl.workMod.NewScratch(token.UserID, name, tool, types, picture, description, toolURL, toolOBJ.Category)

	if err := workCtrl.workMod.RegisteredWork(newWorkContent); err != nil {

		workCtrl.abortWithError(m.ERR_ADD_WORK_FAIL)

	}

	f, _, err := workCtrl.GetFile("file")

	if err != nil {
		log.Fatal("getfile err ", err)
	}

	filePath := "asset/works/" + newWorkContent.ID.Hex() + ".sb3"
	// 保存位置在 static/upload, 没有文件夹要先创建
	if workCtrl.SaveToFile("file", filePath); err != nil {

		workCtrl.abortWithError(m.ERR_ADD_WORK_FAIL)
	}

	defer f.Close()

	out := make(map[string]interface{})
	out["code"] = 0
	workCtrl.jsonResult(out)
}

// LoadProject 获取Scratch作品列表
func (workCtrl *WorksController) LoadProject() {

	token := workCtrl.checkToken()

	userID := token.UserID
	// 获取请求参数中的分页数据
	paging, err := paramPaging(workCtrl.Ctx)

	if err != nil {
		workCtrl.abortWithError(m.ERR_REQUEST_PARAM)
	}

	total, _ := workCtrl.workMod.ScratchProjectCount(userID)

	tool := workCtrl.GetString("tool")

	fmt.Println(paging, userID, total, tool)

	var works []m.WorkListBody

	tool = strings.ToLower(tool)

	works, err = workCtrl.workMod.GetProjectByID(userID, paging)

	fmt.Println(works)

	if err != nil {

		workCtrl.abortWithError(m.ERR_NO_WORK_EXISTS)

	}

	out := make(map[string]interface{})
	out["code"] = 0
	out["works"] = works
	out["total"] = total.Total

	workCtrl.jsonResult(out)

}

//Save3DOne 获取Save3D-One作品列表
//支持3D-ONE 作品上传、下载（上传 STL、封面图片、源文件）， 支持3D渲染预览
func (workCtrl *WorksController) Save3DOne() {
	token := workCtrl.checkToken()
	var workContent m.WorkForm
	if err := workCtrl.ParseForm(&workContent); err != nil {
		workCtrl.abortWithError(m.ERR_REQUEST_PARAM)
	}
	workid := m.NewID()
	relpath := path.Join("asset", "works", workid+".stl")
	stlworkpath := path.Join(beego.AppPath, relpath)
	Z1workpath := path.Join(beego.AppPath, "asset", "works", workid+".Z1")
	dirPath := path.Join(beego.AppPath, "asset", "works")
	stlData, err := workCtrl.GetFiles("stlData")
	if err = UploadFiles(stlworkpath, stlData); err != nil {
		workCtrl.abortWithError(m.ERR_UPLOAD_IMAGES_FAIL)
	}
	Z1Data, err := workCtrl.GetFiles("Z1Data")
	if err = UploadFiles(Z1workpath, Z1Data); err != nil {
		workCtrl.abortWithError(m.ERR_UPLOAD_IMAGES_FAIL)

	}
	if err != nil && "http: no such file" != fmt.Sprintf("%v", err) {
		workCtrl.abortWithError(m.ERR_REQUEST_PARAM)
	}
	logs.Info("workContent.ContentID:", workContent.ContentID)
	newWorkContent := m.NewWork(bson.ObjectIdHex(workid), bson.ObjectIdHex(token.UserID), workContent.ContentID, workContent.Name, workContent.Tool, workContent.Types, relpath, workContent.Picture, workContent.Description, workContent.Data, workContent.ToolURL, workContent.Category, workContent.Public)
	if err := workCtrl.workMod.RegisteredWork(newWorkContent); err != nil {
		logs.Info(err)
		workCtrl.abortWithError(m.ERR_ADD_WORK_FAIL)
	}

	_, e := os.Stat(dirPath)
	if e != nil {
		err := os.Mkdir(dirPath, os.ModePerm)
		if err != nil {
			fmt.Printf("mkdir failed![%v]\n", err)
			workCtrl.abortWithError(m.ERR_CREATE_FILE_FAIL)
		} else {
			fmt.Printf("mkdir success!\n")
		}
	}

	out := make(map[string]interface{})
	out["code"] = 0
	workCtrl.jsonResult(out)
}

//DownloadWork 下载作品源文件或stl文件
func (workCtrl *WorksController) DownloadWork() {
	var id string
	var suffix string
	var filename string
	if err := workCtrl.Ctx.Input.Bind(&id, "id"); err != nil {
		workCtrl.abortWithError(m.ERR_REQUEST_PARAM)
	}
	if err := workCtrl.Ctx.Input.Bind(&suffix, "suffix"); err != nil {
		workCtrl.abortWithError(m.ERR_REQUEST_PARAM)
	}
	if suffix == "Z1" {
		filename = id + ".Z1"
	} else if suffix == "stl" {
		filename = id + ".stl"
	}
	fullPath := path.Join(beego.AppPath, "asset", "works", filename)
	_, err := os.Stat(fullPath)

	if err != nil {

		workCtrl.abortWithError(m.ERR_NO_WORK_EXISTS)

	}
	workCtrl.Ctx.Output.Download(path.Join(beego.AppPath, "asset", "works", filename))
}

//UploadFiles 同时上传数据及文件时multipart
func UploadFiles(filePath string, fileData []*multipart.FileHeader) error {
	// 遍历文件上传到指定位置
	file, err := fileData[0].Open()
	if err != nil {
		logs.Error("err:", err)
		return err
	}
	defer file.Close()
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	io.Copy(f, file)
	return nil
}

//GetClassStudents 获取指定班级学生作品
func (workCtrl *WorksController) GetClassStudents() {
	var classCode string
	var contentID string
	if err := workCtrl.Ctx.Input.Bind(&classCode, "classCode"); err != nil {
		workCtrl.abortWithError(m.ERR_REQUEST_PARAM)
	}
	if err := workCtrl.Ctx.Input.Bind(&contentID, "contentID"); err != nil {
		workCtrl.abortWithError(m.ERR_REQUEST_PARAM)
	}
	if !bson.IsObjectIdHex(contentID) {
		workCtrl.abortWithError(m.ERR_REQUEST_PARAM)
	}
	results, err := workCtrl.workMod.GetClassStudents(classCode, contentID)
	if err != nil {
		workCtrl.abortWithError(m.ERR_READ_WORK_FAIL)
	}
	out := make(map[string]interface{})
	out["code"] = 0
	out["results"] = results
	workCtrl.jsonResult(out)
}
