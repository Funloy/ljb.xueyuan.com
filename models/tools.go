package models

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/mgo.v2/bson"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/cavaliercoder/grab"
	"github.com/gorilla/websocket"
	"github.com/mholt/archiver"
	mgo "gopkg.in/mgo.v2"
	"maiyajia.com/services/mongo"
	"maiyajia.com/util"
)

type ToolModels struct {
	MgoSession *mongo.MgoClient
}

//升级阶段常量
const (
	_             = iota
	DOWNLOAD      //下载阶段
	UNCOMPRESSION //解压阶段
	DOEN          //全部完成
)

type Out struct {
	Code int `json:"code"`
	// Message string    `json:"message"`
	State *WSProgress `json:"state"`
}
type WSProgress struct {
	Stage    int    `json:"stage"`
	Name     string `json:"name"`
	Progress int    `json:"progress"`
}
type ToolListRes struct {
	Code  int    `json:"code"`
	Tools []Tool `json:"tools"`
}
type UpgradeTool struct {
	Code   int    `json:"code"`
	Newver bool   `json:"newver"`
	Tools  []Tool `json:"tools"`
}
type Tool struct {
	ID          bson.ObjectId `bson:"_id" json:"id"`
	DownloadURL string        `bson:"download_url" json:"download_url"`
	Name        string        `bson:"name" json:"name"`
	Relpath     string        `bson:"relpath" json:"relpath"` //course的URL
	Title       string        `bson:"title" json:"title"`
	Category    string        `bson:"category" json:"category"`
	Version     string        `bson:"version" json:"version"`
	Icon        string        `bson:"icon" json:"icon"`
	Purchased   bool          `bson:"purchased" json:"purchased"`
	Weight      int           `bson:"weight" json:"weight"` //权重
	Types       []string      `bson:"types" json:"types"`
	CreateTime  time.Time     `bson:"createTime" json:"createTime"`
}

//Category 工具类别
type Category struct {
	Category string `bson:"category" json:"category"`
	Tools    []Tool `bson:"tools" json:"-"`
	Count    int    `bson:"count" json:"-"`
}

//ToolCategory 工具分类
type ToolCategory struct {
	Name  string `bson:"name" json:"category_name"`
	Tools []Tool `bson:"tools" json:"tools"`
	Count int    `bson:"count" json:"count"`
}

//GetTool 获取单个工具
func (toolMod *ToolModels) GetTool(name string) (Tool, error) {
	var tool Tool
	f := func(col *mgo.Collection) error {
		return col.Find(bson.M{"name": name}).One(&tool)
	}
	err := toolMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "tool", f)
	return tool, err
}

//GetAllToolsByWeight 获取所有工具列表
func (toolMod *ToolModels) GetAllToolsByWeight(paging PagingInfo) ([]Tool, error) {
	var tools []Tool
	offset := paging.Offset()
	limit := paging.Limit()
	f := func(col *mgo.Collection) error {
		return col.Find(nil).Sort("-weight").Limit(limit).Skip(offset).All(&tools)
	}
	err := toolMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "tool", f)
	return tools, err
}

//GetAllTools 获取所有工具列表
func (toolMod *ToolModels) GetAllTools() ([]ToolCategory, error) {
	var tools []ToolCategory
	f := func(col *mgo.Collection) error {
		pipeline := []bson.M{
			{"$group": bson.M{"_id": "$category", "name": bson.M{"$first": "$category"}, "count": bson.M{"$sum": 1}, "tools": bson.M{"$push": "$$ROOT"}}},
			{"$sort": bson.M{"count": -1}},
		}
		return col.Pipe(pipeline).All(&tools)
	}
	err := toolMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "tool", f)
	if err != nil {
		logs.Error("Begin getTools in models:", err)
	}
	return tools, err
}

//GetToolsCategory 获取所有工具类别
func (toolMod *ToolModels) GetToolsCategory() ([]Category, error) {
	var categorys []Category
	f := func(col *mgo.Collection) error {
		pipeline := []bson.M{
			{"$group": bson.M{"_id": "$category", "category": bson.M{"$first": "$category"}, "count": bson.M{"$sum": 1}, "tools": bson.M{"$push": "$$ROOT"}}},
			{"$sort": bson.M{"count": -1}},
		}
		return col.Pipe(pipeline).All(&categorys)
	}
	err := toolMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "tool", f)
	if err != nil {
		logs.Error("Begin getTools in models:", err)
	}
	return categorys, err
}

//GetToolsInfo 获取所有工具信息，用于升级新增
func (toolMod *ToolModels) GetToolsInfo() (interface{}, error) {
	var tools []bson.M
	f := func(col *mgo.Collection) error {
		pipeline := []bson.M{
			{"$project": bson.M{"name": 1, "version": 1}},
		}
		return col.Pipe(pipeline).All(&tools)
	}
	err := toolMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "tool", f)
	if err != nil {
		logs.Error("Begin getTools in models:", err)
	}
	return tools, err
}

//InstallTools 接口安装
func (toolMod *ToolModels) InstallTools(url string, productKey string, productSerial string, ws *websocket.Conn) error {
	logs.Info("begin Install tool")
	mongo.Client = &mongo.MgoClient{}
	mongo.Client.StartSession()
	defer mongo.Client.CloseSession()
	//1. 调用课程平台queryList接口获取所有的课程信息
	var tools ToolListRes

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logs.Error("NewRequest erro", err)
		return err
	}
	req.Header.Set("productKey", productKey)
	req.Header.Set("productSerial", productSerial)
	resp, err := client.Do(req)
	if err != nil {
		logs.Error("resp Error", err)
		return err
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	toolJSON := buf.String()
	logs.Info("toolJSON:", toolJSON)
	err = json.Unmarshal([]byte(toolJSON), &tools)
	if err != nil {
		logs.Error("unmarshal json err:", err)
		return err
	}
	//清空数据库工具信息
	if err := toolMod.clearToolData(); err != nil {
		logs.Info("clearToolData error:", err)
		return err
	}
	//多文件下载
	if err := toolMod.DownloadTools(tools.Tools, ws); err != nil {
		logs.Error("DownloadTools is err:", err)
		return err
	}
	logs.Info("end Install tool")
	return nil
}

//DownloadTools WS多文件下载
func (toolMod *ToolModels) DownloadTools(tools []Tool, ws *websocket.Conn) error {
	var out Out
	logs.Info("Begin Download Tools")
	if tools == nil {
		logs.Info("Tools is null")
		return fmt.Errorf("Tools is nil")
	}
	reqs := make([]*grab.Request, 0)
	//判断是否为七牛云公开下载
	isQINIU := beego.AppConfig.DefaultBool("isQINIU", false)
	if isQINIU {
		for _, tool := range tools {
			// if tool.Purchased == false {
			// 	continue
			// }
			logs.Info("工具名字：", tool.Name)
			req, _ := grab.NewRequest(path.Join(beego.AppPath, "asset", "tools", tool.Name, "tool.tar"), tool.DownloadURL)
			logs.Info("tool.DownloadURL:", tool.DownloadURL)
			req.Tag = tool
			reqs = append(reqs, req)
		}
	} else {
		token, err := util.NetdiskAuth()
		if err != nil {
			logs.Error("网盘登陆失败：", err)
			return err
		}
		for _, tool := range tools {
			//判断是否购买
			// if tool.Purchased == false {
			// 	continue
			// }
			logs.Info("工具名字：", tool.Name)
			req, _ := grab.NewRequest(path.Join(beego.AppPath, "asset", "tools", tool.Name, "tool.tar"), tool.DownloadURL)
			req.HTTPRequest.Header.Add("Authorization", "Bearer "+string(token))
			req.Tag = tool
			reqs = append(reqs, req)
		}
	}
	// start files downloads, arg0 at a time
	respch := grab.DefaultClient.DoBatch(0, reqs...)
	t := time.NewTicker(200 * time.Millisecond)
	// monitor downloads
	completed := 0
	inProgress := 0
	responses := make([]*grab.Response, 0)
	for completed < len(reqs) {
		select {
		case resp := <-respch:
			// a new response has been received and has started downloading
			// (nil is received once, when the channel is closed by grab)
			if resp != nil {
				responses = append(responses, resp)
			}

		case <-t.C:
			// update completed downloads
			inProgress = 0
			for i, resp := range responses {
				// update downloads in progress
				if resp != nil {
					inProgress++
					out.Code = 0
					out.State = &WSProgress{
						Stage:    DOWNLOAD,
						Name:     resp.Request.Tag.(Tool).Name,
						Progress: int(100 * resp.Progress()),
					}
					send_msg, err := json.Marshal(out)
					if err != nil {
						logs.Error("error:", err)
					}
					ws.WriteMessage(websocket.TextMessage, []byte(send_msg))
					logs.Info("Transferred %v/%v bytes (%.2f)", resp.BytesComplete(), resp.Size, 100*resp.Progress())
				}
				if resp != nil && resp.IsComplete() {
					////提示下载完成
					out.Code = 0
					out.State = &WSProgress{
						Stage:    DOWNLOAD,
						Name:     resp.Request.Tag.(Tool).Name,
						Progress: 100,
					}
					send_msg, err := json.Marshal(out)
					ws.WriteMessage(websocket.TextMessage, []byte(send_msg))
					if err != nil {
						logs.Error("error:", err)
					}
					// print final result
					tool := resp.Request.Tag.(Tool)
					if resp.Err() != nil {
						logs.Info("Download failed: %v\n", resp.Err())
						return resp.Err()
					} else {
						logs.Info("Download saved to ./%v \n", resp.Filename)
						//Open解压文件到指定文件夹asset
						// 解压安装包
						out.Code = 0
						out.State = &WSProgress{
							Stage:    UNCOMPRESSION,
							Name:     resp.Request.Tag.(Tool).Name,
							Progress: 0,
						}
						send_msg, err := json.Marshal(out)
						if err != nil {
							logs.Error("error:", err)
						}
						ws.WriteMessage(websocket.TextMessage, []byte(send_msg))
						if error := archiver.Tar.Open(path.Join(beego.AppPath, "asset", "tools", tool.Name, "tool.tar"), path.Join(beego.AppPath, "asset", "tools", tool.Name)); error != nil {
							logs.Info("File extract fail!", error)
							return error
						}
						if err := toolMod.handleTool(tool); err != nil {
							return err
						}
						// 解压安装包
						out.Code = 0
						out.State = &WSProgress{
							Stage:    UNCOMPRESSION,
							Name:     resp.Request.Tag.(Tool).Name,
							Progress: 100,
						}
						send_msg, err = json.Marshal(out)
						ws.WriteMessage(websocket.TextMessage, []byte(send_msg))
					}
					// mark completed
					responses[i] = nil
					completed++
				}
			}
		}

	}
	out.Code = 0
	out.State = &WSProgress{
		Stage: DOEN,
	}
	send_msg, _ := json.Marshal(out)
	ws.WriteMessage(websocket.TextMessage, []byte(send_msg))
	t.Stop()
	logs.Info("%d files successfully downloaded.\n", completed)
	return nil
}

//OrderInstallTools 指令安装
func (toolMod *ToolModels) OrderInstallTools(url string, productKey string, productSerial string) error {
	logs.Info("begin Install tool")
	mongo.Client = &mongo.MgoClient{}
	mongo.Client.StartSession()
	defer mongo.Client.CloseSession()
	//1. 调用课程平台queryList接口获取所有的课程信息
	var tools ToolListRes

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logs.Error("NewRequest erro", err)
		return err
	}
	req.Header.Set("productKey", productKey)
	req.Header.Set("productSerial", productSerial)
	resp, err := client.Do(req)
	if err != nil {
		logs.Error("resp Error", err)
		return err
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	toolJSON := buf.String()
	logs.Info("toolJSON:", toolJSON)
	err = json.Unmarshal([]byte(toolJSON), &tools)
	if err != nil {
		logs.Error("unmarshal json err:", err)
		return err
	}
	//清空数据库工具信息
	if err := toolMod.clearToolData(); err != nil {
		logs.Info("clearToolData error:", err)
		return err
	}
	//多文件下载
	if err := toolMod.OrderDownloadTools(tools.Tools); err != nil {
		logs.Error("DownloadTools is err:", err)
		return err
	}
	logs.Info("end Install tool")
	return nil
}

//OrderDownloadTools 后台指令多文件下载
func (toolMod *ToolModels) OrderDownloadTools(tools []Tool) error {
	logs.Info("Begin Download Tools")
	if tools == nil {
		logs.Info("Tools is null")
		return fmt.Errorf("Tools is nil")
	}
	reqs := make([]*grab.Request, 0)
	//判断是否为七牛云公开下载
	isQINIU := beego.AppConfig.DefaultBool("isQINIU", false)
	if isQINIU {
		for _, tool := range tools {
			// if tool.Purchased == false {
			// 	continue
			// }
			logs.Info("工具名字：", tool.Name)
			req, _ := grab.NewRequest(path.Join(beego.AppPath, "asset", "tools", tool.Name, "tool.tar"), tool.DownloadURL)
			logs.Info("tool.DownloadURL:", tool.DownloadURL)
			req.Tag = tool
			reqs = append(reqs, req)
		}
	} else {
		token, err := util.NetdiskAuth()
		if err != nil {
			logs.Error("网盘登陆失败：", err)
			return err
		}
		for _, tool := range tools {
			//判断是否购买
			// if tool.Purchased == false {
			// 	continue
			// }
			logs.Info("工具名字：", tool.Name)
			req, _ := grab.NewRequest(path.Join(beego.AppPath, "asset", "tools", tool.Name, "tool.tar"), tool.DownloadURL)
			req.HTTPRequest.Header.Add("Authorization", "Bearer "+string(token))
			req.Tag = tool
			reqs = append(reqs, req)
		}
	}
	// start files downloads, arg0 at a time
	respch := grab.DefaultClient.DoBatch(0, reqs...)
	t := time.NewTicker(200 * time.Millisecond)
	// monitor downloads
	completed := 0
	inProgress := 0
	responses := make([]*grab.Response, 0)
	for completed < len(reqs) {
		select {
		case resp := <-respch:
			// a new response has been received and has started downloading
			// (nil is received once, when the channel is closed by grab)
			if resp != nil {
				responses = append(responses, resp)
			}

		case <-t.C:
			// update completed downloads
			inProgress = 0
			for i, resp := range responses {
				// update downloads in progress
				if resp != nil {
					inProgress++
					logs.Info("Transferred %v/%v bytes (%.2f)", resp.BytesComplete(), resp.Size, 100*resp.Progress())
				}
				if resp != nil && resp.IsComplete() {
					// print final result
					tool := resp.Request.Tag.(Tool)
					if resp.Err() != nil {
						logs.Info("Download failed: %v\n", resp.Err())
						return resp.Err()
					} else {
						logs.Info("Download saved to ./%v \n", resp.Filename)
						//Open解压文件到指定文件夹asset
						if error := archiver.Tar.Open(path.Join(beego.AppPath, "asset", "tools", tool.Name, "tool.tar"), path.Join(beego.AppPath, "asset", "tools", tool.Name)); error != nil {
							logs.Info("File extract fail!", error)
							return error
						}
						if err := toolMod.handleTool(tool); err != nil {
							return err
						}
					}
					// mark completed
					responses[i] = nil
					completed++
				}
			}
		}

	}
	t.Stop()
	logs.Info("%d files successfully downloaded.\n", completed)
	return nil
}

/*********************************************************************************************/
/*********************************** 以下为本控制器的内部函数 *********************************/
/*********************************** *********************************************************/
//获取图标路径
func getIconURL(indexpath string, tool *Tool) {
	filepng, _ := ioutil.ReadDir(indexpath)
	files, _ := ioutil.ReadDir(path.Join(indexpath, filepng[0].Name()))
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".svg") {
			logs.Info("tool.Icon:=", f.Name())
			tool.Icon = "asset/tools/" + tool.Name + "/" + filepng[0].Name() + "/" + f.Name()
			break
		}
	}
}

//InsertTool 写入下载工具信息
func (toolMod *ToolModels) insertTool(tool Tool) error {
	f := func(col *mgo.Collection) error {
		query := bson.M{"name": tool.Name}
		change := mgo.Change{
			//Update:    bson.M{"$set": bson.M{"name": account.Name, "key": account.Key, "serial": account.Serial, "os": account.OS, "version": account.Version, "newver": account.Newver, "upgrade": account.Upgrade}},
			Update:    tool,
			ReturnNew: true,
			Upsert:    true,
		}
		_, err := col.Find(query).Apply(change, nil)
		return err
	}
	return toolMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "tool", f)
}

//getFilePath 递归获取文件相对路径
func getFilePath(indexpath string, relpath *string) error {
	files, _ := ioutil.ReadDir(indexpath)
	for _, f := range files {
		if f.IsDir() {
			getFilePath(path.Join(indexpath, f.Name()), relpath)
		} else if "index.html" == f.Name() {
			path, err := filepath.Rel(beego.AppPath, indexpath)
			if err != nil {
				logs.Error("get relpath fail")
				return err
			}
			*relpath = path + "/index.html"
			break
		}
	}

	return nil
}

//handleTool 写入数据库前对tool进行处理
func (toolMod *ToolModels) handleTool(tool Tool) error {
	logs.Info("Tool_file extract done")
	if error := os.Remove(path.Join(beego.AppPath, "asset", "tools", tool.Name, "tool.tar")); error != nil {
		//如果删除失败则输出 file remove Error!
		logs.Info("zip file remove Error!")
		return error
	} else {
		//如果删除成功则输出 file remove OK!
		logs.Info("zip file remove OK!")
	}
	var relpath string
	indexpath := path.Join(beego.AppPath, "asset", "tools", tool.Name)
	getIconURL(indexpath, &tool)
	getFilePath(indexpath, &relpath)
	filepath := strings.Replace(relpath, "\\", "/", -1)
	tool.Relpath = filepath
	//写入工具信息
	if err := toolMod.insertTool(tool); err != nil {
		logs.Info("insertTool fail:", err)
	}

	return nil
}
func (toolMod *ToolModels) clearToolData() error {
	f := func(col *mgo.Collection) error {
		_, err := col.RemoveAll(bson.M{})
		return err
	}
	toolMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "tool", f)
	return nil
}
