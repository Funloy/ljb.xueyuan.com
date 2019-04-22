package models

import (
	"fmt"
	"io/ioutil"
	"path"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"maiyajia.com/services/mongo"
)

type WorkModels struct {
	MgoSession *mongo.MgoClient
}

//bson.ObjectIdHex()
//bson.ObjectId
const (
	_          = iota
	BROWSE     //浏览量排
	CREATETIME //创建时间排序
)

type BaseBody struct {
	ID          bson.ObjectId `bson:"_id" json:"id"`                  //作品ID
	UserID      bson.ObjectId `bson:"userID" json:"userID"`           //用户
	Name        string        `bson:"name" json:"name"`               //作品名称
	Picture     string        `bson:"picture" json:"picture"`         //作品封面
	Laud        int64         `bson:"laud" json:"laud"`               //作品点赞数
	Favor       int64         `bson:"favor" json:"favor"`             //作品收藏数
	Browse      int64         `bson:"browse" json:"browse"`           //作品浏览数
	Tool        string        `bson:"tool" json:"tool"`               //默认打开的工具
	Description string        `bson:"description" json:"description"` //作品简介
	Edit        bool          `bson:"edit" json:"edit"`               //分享后是否允许编辑
	Types       string        `bson:"types" json:"types"`             //作品类型
	ContentID   bson.ObjectId `bson:"contentID" json:"contentID"`     //课节ID
}

// WorkBody 作品数据结构
type WorkBody struct {
	BaseBody   `bson:",inline"`
	OriginID   string `bson:"originID" json:"-"`            //原作品id(用于收藏)
	Data       string `bson:"data" json:"data"`             //作品内容
	Relpath    string `bson:"relpath" json:"relpath"`       //作品的下载地址
	ToolURL    string `bson:"toolURL" json:"toolURL"`       //下载地址
	CreateTime int64  `bson:"createTime" json:"createTime"` //创建时间
	Public     bool   `bson:"public" json:"public"`         //是否分享true分享
	Category   string `bson:"category" json:"category"`
}
type WorkForm struct {
	ID          bson.ObjectId `bson:"_id" form:"id"`                  //作品ID
	UserID      bson.ObjectId `bson:"userID" form:"userID"`           //用户
	Name        string        `bson:"name" form:"name"`               //作品名称
	Picture     string        `bson:"picture" form:"picture"`         //作品封面
	Laud        int64         `bson:"laud" form:"laud"`               //作品点赞数
	Favor       int64         `bson:"favor" form:"favor"`             //作品收藏数
	Browse      int64         `bson:"browse" form:"browse"`           //作品浏览数
	Tool        string        `bson:"tool" form:"tool"`               //默认打开的工具
	Description string        `bson:"description" form:"description"` //作品简介
	Edit        bool          `bson:"edit" form:"edit"`               //分享后是否允许编辑
	OriginID    string        `bson:"originID" form:"-"`              //原作品id(用于收藏)
	Types       string        `bson:"types" form:"types"`             //作品类型
	Data        string        `bson:"data" form:"data"`               //作品内容
	Relpath     string        `bson:"relpath" form:"relpath"`         //作品的下载地址
	ToolURL     string        `bson:"toolURL" form:"toolURL"`         //下载地址
	CreateTime  int64         `bson:"createTime" form:"createTime"`   //创建时间
	Public      bool          `bson:"public" form:"public"`           //是否分享true分享
	Category    string        `bson:"category" form:"category"`
	ContentID   string        `bson:"contentID" form:"contentID"`
}

// WorkListBody 作品列表数据结构
type WorkListBody struct {
	BaseBody   `bson:",inline"`
	RealName   string `bson:"realname" json:"realname"`     //用户名
	CreateTime int64  `bson:"createTime" json:"createTime"` //创建时间
	Data       string `bson:"data" json:"data"`             //作品内容
	ToolURL    string `bson:"toolURL" json:"toolURL"`       //下载地址
}

//RecordFavorOrLaud 收藏、点赞记录
type RecordFavorOrLaud struct {
	ID     bson.ObjectId `bson:"_id" json:"id"`
	WorkID bson.ObjectId `bson:"workid" json:"workid"`
	UserID bson.ObjectId `bson:"userid" json:"userid"`
}
type TotalBody struct {
	Total int `bson:"total" json:"total"`
}
type ActiveCount struct {
	// Year  int `bson:"year" json:"year"`
	// Month int `bson:"month" json:"month"`
	Time  time.Time `bson:"time" json:"time"`
	Total int       `bson:"total" json:"total"`
}

//PerShareCou 个人作品分享、被收藏数
type PerShareCou struct {
	Favor int `bson:"favor" json:"favor"` //作品收藏数
	Laud  int `bson:"laud" json:"laud"`   //作品点赞数
}

//WorksCategoryByTool 按工具分类展示作品
type WorksCategoryByTool struct {
	Category string         `bson:"category" json:"category"`
	Works    []WorkListBody `bson:"works" json:"works"`
	Count    int            `bson:"count" json:"count"`
}

//ShareWork 作品分享属性
type ShareWork struct {
	ID     bson.ObjectId `bson:"_id" json:"id"`
	Public bool          `bson:"public" json:"public"`
	Edit   bool          `bson:"edit" json:"edit"` //分享后是否允许编辑
}

type DataBody struct {
	X int `bson:"x" json:"x"`
	Y int `bson:"y" json:"y"`
}

type Data struct {
	ID string `json:"id"`
}

// PaginationBody 分页数据结构
type PaginationBody struct {
	Total    int `bson:"total" json:"total"`
	PageSize int `bson:"pageSize" json:"pageSize"`
	Current  int `bson:"current" json:"current"`
}

// NewID 创建新的ID
func NewID() string {

	return bson.NewObjectId().Hex()
}

func NewIDBson() bson.ObjectId {

	return bson.NewObjectId()
}

// GetDesc 获取指定ID的作品
func (workMod *WorkModels) GetDesc(id string) (WorkBody, error) {
	var work WorkBody
	f := func(col *mgo.Collection) error {
		return col.Find(bson.M{"_id": bson.ObjectIdHex(id)}).Select(bson.M{"_id": 1, "name": 1, "picture": 1, "description": 1, "createTime": 1, "data": 1, "tool": 1}).One(&work)
	}
	err := workMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "works", f)
	return work, err
}

// GetPreview 获取预览作品信息
func (workMod *WorkModels) GetPreview(id string) (map[string]interface{}, error) {
	work := make(map[string]interface{})
	pipeline := []bson.M{
		{"$match": bson.M{"_id": bson.ObjectIdHex(id)}},
		{"$lookup": bson.M{
			"from": "users",
			"let":  bson.M{"userID": "$userID"},
			"pipeline": []bson.M{
				{"$match": bson.M{"$expr": bson.M{"$eq": []interface{}{"$_id", "$$userID"}}}},
			},
			"as": "user",
		},
		},
		{"$project": bson.M{
			"_id":         0,
			"id":          "$_id",
			"userID":      1,
			"name":        1,
			"picture":     1,
			"toolURL":     1,
			"description": 1,
			"createTime":  1,
			"browse":      1,
			"favor":       1,
			"laud":        1,
			"types":       1,
			"edit":        1,
			"contentID":   1,
			"username":    "$user.username",
			"avatar":      "$user.avatar",
		}},
	}
	f := func(col *mgo.Collection) error {
		return col.Pipe(pipeline).One(&work)
	}

	err := workMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "works", f)

	return work, err

}

//GetWorksByCategory 获取分享的作品列表
func (workMod *WorkModels) GetWorksByCategory(paging PagingInfo, sortCategory int, category string) (interface{}, error) {
	//var works []WorkListBody
	var works []bson.M
	var sort string

	if sortCategory == CREATETIME {
		sort = "createTime"
	} else {
		sort = "browse"
	}
	offset := paging.Offset()
	limit := paging.Limit()

	pipeline := []bson.M{
		{"$match": bson.M{"public": true, "category": category}},
		{"$lookup": bson.M{
			"from":         "users",
			"foreignField": "_id",
			"localField":   "userID",
			"as":           "user",
		}},
		{"$project": bson.M{
			"_id":           0,
			"id":            "$_id",
			"userID":        1,
			"name":          1,
			"picture":       1,
			"description":   1,
			"createTime":    1,
			"laud":          1,
			"browse":        1,
			"tool":          1,
			"toolURL":       1,
			"favor":         1,
			"edit":          1,
			"user.realname": 1,
		},
		},
		{"$sort": bson.M{sort: -1}},
		{"$skip": offset},
		{"$limit": limit},
	}
	f := func(col *mgo.Collection) error {
		return col.Pipe(pipeline).All(&works)
	}
	// f := func(col *mgo.Collection) error {
	// 	return col.Find(bson.M{"public": true, "category": category}).Sort(sort).Limit(limit).Skip(offset).All(&works)
	// }
	err := workMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "works", f)
	return works, err
}

// GetPersonShareWorks 获取个人分享的作品列表
func (workMod *WorkModels) GetPersonShareWorks(paging PagingInfo, userID string) ([]WorkListBody, error) {

	offset := paging.Offset()
	limit := paging.Limit()

	var works []WorkListBody
	f := func(col *mgo.Collection) error {
		return col.Find(bson.M{"public": true, "userID": bson.ObjectIdHex(userID)}).Select(bson.M{"_id": 1, "userID": 1, "name": 1, "picture": 1, "description": 1, "createTime": 1, "laud": 1, "browse": 1, "tool": 1, "toolURL": 1, "favor": 1, "edit": 1}).Limit(limit).Skip(offset).All(&works)
	}

	err := workMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "works", f)

	return works, err

}

//GetPersonShareWorksCou 获取个人分享的作品
func (workMod *WorkModels) GetPersonShareWorksCou(userID string) ([]PerShareCou, error) {
	var perShareCous []PerShareCou

	f := func(col *mgo.Collection) error {
		return col.Find(bson.M{"public": true, "userID": bson.ObjectIdHex(userID)}).Select(bson.M{"laud": 1, "favor": 1}).All(&perShareCous)
	}
	err := workMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "works", f)
	return perShareCous, err
}

// GetShareWorks 获取分享的作品列表
func (workMod *WorkModels) GetShareWorks(paging PagingInfo, sortCategory int) (interface{}, error) {

	offset := paging.Offset()
	limit := paging.Limit()
	var sort string
	if sortCategory == CREATETIME {
		sort = "createTime"
	} else {
		sort = "browse"
	}
	var works []bson.M
	// f := func(col *mgo.Collection) error {
	// 	return col.Find(bson.M{"public": true}).Select(bson.M{"_id": 1, "userID": 1, "name": 1, "picture": 1, "description": 1, "createTime": 1, "laud": 1, "browse": 1, "tool": 1, "toolURL": 1, "favor": 1}).Sort(sort).Limit(limit).Skip(offset).All(&works)
	// }
	pipeline := []bson.M{
		{"$match": bson.M{"public": true}},
		{"$lookup": bson.M{
			"from":         "users",
			"foreignField": "_id",
			"localField":   "userID",
			"as":           "user",
		}},
		{"$project": bson.M{
			"_id":           0,
			"id":            "$_id",
			"userID":        1,
			"name":          1,
			"picture":       1,
			"description":   1,
			"createTime":    1,
			"laud":          1,
			"browse":        1,
			"tool":          1,
			"toolURL":       1,
			"favor":         1,
			"edit":          1,
			"user.realname": 1,
		}},
		{"$sort": bson.M{sort: -1}},
		{"$skip": offset},
		{"$limit": limit},
	}
	f := func(col *mgo.Collection) error {
		return col.Pipe(pipeline).All(&works)
	}
	err := workMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "works", f)

	return works, err

}

// GetWorksByID 获取指定用户ID的所有作品
func (workMod *WorkModels) GetWorksByID(userID string, params string, paging PagingInfo) ([]WorkListBody, error) {

	offset := paging.Offset()
	limit := paging.Limit()

	var works []WorkListBody

	f := func(col *mgo.Collection) error {
		return col.Find(bson.M{"userID": bson.ObjectIdHex(userID)}).Select(bson.M{"_id": 1, "name": 1, "picture": 1, "description": 1, "createTime": 1, "data": 1, "tool": 1, "public": 1, "toolURL": 1, "types": 1}).Limit(limit).Skip(offset).All(&works)
	}
	if params != "" {
		f = func(col *mgo.Collection) error {
			return col.Find(bson.M{"userID": bson.ObjectIdHex(userID), "tool": params}).Select(bson.M{"_id": 1, "name": 1, "picture": 1, "description": 1, "createTime": 1, "data": 1, "tool": 1, "public": 1, "toolURL": 1, "types": 1}).Limit(limit).Skip(offset).All(&works)
		}
	}

	err := workMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "works", f)

	return works, err

}

// QueryShareWorksCount 查询作品分享次数
func (workMod *WorkModels) QueryShareWorksCount() (TotalBody, error) {

	pipeline := []bson.M{
		{"$match": bson.M{"public": true}},
		{"$group": bson.M{"_id": nil, "total": bson.M{"$sum": 1}}},
		{"$project": bson.M{"_id": 0}},
	}
	var total TotalBody
	f := func(col *mgo.Collection) error {
		return col.Pipe(pipeline).One(&total)
	}
	return total, workMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "works", f)
}

// QueryPersonShareWorksCount 查询个人作品分享次数
func (workMod *WorkModels) QueryPersonShareWorksCount(userID string) (TotalBody, error) {

	pipeline := []bson.M{
		{"$match": bson.M{"public": true, "userID": bson.ObjectIdHex(userID)}},
		{"$group": bson.M{"_id": nil, "total": bson.M{"$sum": 1}}},
		{"$project": bson.M{"_id": 0}},
	}
	// result := bson.M{}
	var total TotalBody
	f := func(col *mgo.Collection) error {
		return col.Pipe(pipeline).One(&total)
	}
	return total, workMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "works", f)
}

// QueryWorksByCategoryCount 查询指定类别下
func (workMod *WorkModels) QueryWorksByCategoryCount(category string) (TotalBody, error) {

	pipeline := []bson.M{
		{"$match": bson.M{"public": true, "category": category}},
		{"$group": bson.M{"_id": nil, "total": bson.M{"$sum": 1}}},
		{"$project": bson.M{"_id": 0}},
	}
	var total TotalBody
	f := func(col *mgo.Collection) error {
		return col.Pipe(pipeline).One(&total)
	}
	return total, workMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "works", f)
}

//QueryWorksCount
func (workMod *WorkModels) QueryWorksCount(userID, tableName string) (TotalBody, error) {

	pipeline := []bson.M{
		{"$match": bson.M{"userID": bson.ObjectIdHex(userID)}},
		{"$group": bson.M{"_id": nil, "total": bson.M{"$sum": 1}}},
		{"$project": bson.M{"_id": 0}},
	}
	var total TotalBody
	f := func(col *mgo.Collection) error {
		return col.Pipe(pipeline).One(&total)
	}
	return total, workMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), tableName, f)
}

// FindWorkByID 获取指定作品ID的内容
func (workMod *WorkModels) FindWorkByID(workID bson.ObjectId) (WorkBody, error) {

	var work WorkBody
	f := func(col *mgo.Collection) error {
		return col.Find(bson.M{"_id": workID}).One(&work)
	}
	err := workMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "works", f)
	return work, err

}

// CopyWork 复制指定作品数据
func (workMod *WorkModels) CopyWork(work WorkBody, userID, id, name string, newid bson.ObjectId, toolNames []string) error {
	work.ID = newid
	work.OriginID = id
	work.Name = name
	work.UserID = bson.ObjectIdHex(userID)
	work.Laud = 0
	work.Favor = 0
	work.Browse = 0
	for _, toolName := range toolNames {
		if work.Tool == toolName {
			work.Relpath = ""
		} else {
			work.Relpath = fmt.Sprintf("asset/works/%v.stl", newid.Hex())
		}
	}
	work.Public = false
	work.CreateTime = time.Now().Unix()
	f := func(col *mgo.Collection) error {
		return col.Insert(work)
	}
	return workMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "works", f)

}

// NewWork 初始化新的作品信息
func NewWork(id, userID bson.ObjectId, contentID, name, tool, types, relpath, picture, description, data, toolurl, category string) *WorkBody {
	logs.Info("5c36f60688219d2a7f323b80", contentID)
	if !bson.IsObjectIdHex(contentID) {
		contentID = NewID()
	}
	work := &WorkBody{
		BaseBody: BaseBody{
			ID:          id,
			UserID:      userID,
			Name:        name,
			Tool:        tool,
			Picture:     picture,
			Description: description,
			Types:       types,
			ContentID:   bson.ObjectIdHex(contentID),
		},
		Relpath:    relpath,
		ToolURL:    toolurl,
		CreateTime: time.Now().Unix(),
		Data:       data,
		Category:   category,
	}
	return work
}

// PublicWork 分享作品
func (workMod *WorkModels) PublicWork(id bson.ObjectId, userid string, public, edit bool) error {

	f := func(col *mgo.Collection) error {
		return col.Update(bson.M{"_id": id, "userID": bson.ObjectIdHex(userid)}, bson.M{"$set": bson.M{"public": public, "edit": edit}})
	}
	return workMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "works", f)
}

// PatchWork 更新作品描述
func (workMod *WorkModels) PatchWork(id string, userID bson.ObjectId, name string, tool string, types string, picture string, description string, data string) error {

	f := func(col *mgo.Collection) error {
		return col.Update(bson.M{"_id": bson.ObjectIdHex(id)}, bson.M{"$set": bson.M{"name": name, "userID": userID, "tool": tool, "types": types, "picture": picture, "description": description, "data": data}})
	}
	return workMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "works", f)
}

// RegisteredWork 在数据库创建一个新的作品
func (workMod *WorkModels) RegisteredWork(work interface{}) error {
	f := func(col *mgo.Collection) error {
		return col.Insert(work)
	}
	return workMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "works", f)
}

// DeleteWorkByID 删除指定ID的作品
func (workMod *WorkModels) DeleteWorkByID(workID string) error {

	f := func(col *mgo.Collection) error {
		return col.Remove(bson.M{"_id": bson.ObjectIdHex(workID)})
	}
	err := workMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "works", f)
	return err
}

// DeleteFavorWorkByID 删除作品时同时删除收藏作评的作品
func (workMod *WorkModels) DeleteFavorWorkByID(workID string) error {

	f := func(col *mgo.Collection) error {
		return col.Remove(bson.M{"workid": bson.ObjectIdHex(workID)})
	}
	err := workMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "favor", f)
	return err
}

// FindUserWorkFavor 查询作品的收藏
func (workMod *WorkModels) FindUserWorkFavor(userid string, paging PagingInfo) []*RecordFavorOrLaud {
	var isFavor []*RecordFavorOrLaud
	offset := paging.Offset()
	limit := paging.Limit()
	f := func(col *mgo.Collection) error {
		return col.Find(bson.M{"userid": bson.ObjectIdHex(userid)}).Limit(limit).Skip(offset).All(&isFavor)
	}
	workMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "favor", f)
	return isFavor
}

// FindWorkFavor 查询作品的收藏
func (workMod *WorkModels) FindWorkFavor(workid string) []*RecordFavorOrLaud {
	var isFavor []*RecordFavorOrLaud
	f := func(col *mgo.Collection) error {
		return col.Find(bson.M{"workid": bson.ObjectIdHex(workid)}).All(&isFavor)
	}
	workMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "favor", f)
	return isFavor
}

// GetdataFile 获取存储在文件的作品数据
func GetdataFile(path string) (string, error) {

	b, err := ioutil.ReadFile(path)

	return string(b), err
}

// WritdataFile 存储作品数据
func WritdataFile(path string, data string) error {

	err := ioutil.WriteFile(path, []byte(data), 0666)

	return err
}

//GiveLaud 点赞作品
func (workMod *WorkModels) GiveLaud(id bson.ObjectId) error {
	f := func(col *mgo.Collection) error {
		query := bson.M{"_id": id}
		change := mgo.Change{
			Update:    bson.M{"$inc": bson.M{"laud": 1}},
			ReturnNew: true,
		}
		_, err := col.Find(query).Apply(change, nil)
		return err
	}
	err := workMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "works", f)
	return err
}

//FavorCount 收藏数
func (workMod *WorkModels) FavorCount(id bson.ObjectId) error {
	f := func(col *mgo.Collection) error {
		query := bson.M{"_id": id}
		change := mgo.Change{
			Update:    bson.M{"$inc": bson.M{"favor": 1}},
			ReturnNew: true,
		}
		_, err := col.Find(query).Apply(change, nil)
		return err
	}
	err := workMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "works", f)
	return err
}

// FindIsLaud 查询用户是否点赞
func (workMod *WorkModels) FindIsLaud(workid, userid bson.ObjectId) (interface{}, error) {
	isLaud := bson.M{}
	f := func(col *mgo.Collection) error {
		return col.Find(bson.M{"workid": workid, "userid": userid}).One(&isLaud)

	}

	err := workMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "laud", f)

	return isLaud, err

}

// FindIsFacvor 查询用户是否收藏
func (workMod *WorkModels) FindIsFacvor(workid, userid bson.ObjectId) (interface{}, error) {
	isFavor := bson.M{}
	f := func(col *mgo.Collection) error {
		return col.Find(bson.M{"workid": workid, "userid": userid}).One(&isFavor)

	}

	err := workMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "favor", f)

	return isFavor, err

}

// InsertLaud 用户点赞记录
func (workMod *WorkModels) InsertLaud(recordLaud RecordFavorOrLaud) error {
	f := func(col *mgo.Collection) error {
		return col.Insert(recordLaud)
	}
	return workMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "laud", f)
}

// InsertFavor 用户收藏记录
func (workMod *WorkModels) InsertFavor(recordFavor RecordFavorOrLaud) error {
	f := func(col *mgo.Collection) error {
		return col.Insert(recordFavor)
	}
	return workMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "favor", f)
}

//GetUserFavoriteWork 获取个人收藏索引
func (workMod *WorkModels) GetUserFavoriteWork(workid bson.ObjectId) (bson.M, error) {
	result := bson.M{}
	pipeline := []bson.M{
		{"$match": bson.M{"_id": workid}},
		{"$lookup": bson.M{
			"from":         "users",
			"localField":   "userID",
			"foreignField": "_id",
			"as":           "user",
		},
		},
		{"$project": bson.M{
			"_id":           0,
			"id":            "$_id",
			"userID":        1,
			"name":          1,
			"picture":       1,
			"toolURL":       1,
			"description":   1,
			"createTime":    1,
			"browse":        1,
			"favor":         1,
			"laud":          1,
			"edit":          1,
			"user.realname": 1,
		},
		},
	}
	f := func(col *mgo.Collection) error {
		return col.Pipe(pipeline).One(&result)
	}
	return result, workMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "works", f)
}

// DeleteFavoriteWork 删除个人收藏的作品
func (workMod *WorkModels) DeleteFavoriteWork(workid, userid string) error {

	f := func(col *mgo.Collection) error {
		return col.Remove(bson.M{"workid": bson.ObjectIdHex(workid), "userid": bson.ObjectIdHex(userid)})
	}
	err := workMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "favor", f)
	return err
}

//RecordBrowse 记录作品浏览量
func (workMod *WorkModels) RecordBrowse(workid string) error {
	f := func(col *mgo.Collection) error {
		query := bson.M{"_id": bson.ObjectIdHex(workid)}
		change := mgo.Change{
			Update:    bson.M{"$inc": bson.M{"browse": 1}},
			ReturnNew: true,
		}
		_, err := col.Find(query).Apply(change, nil)
		return err
	}
	err := workMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "works", f)
	return err
}

func (workMod *WorkModels) ScratchProjectCount(userID string) (TotalBody, error) {

	pipeline := []bson.M{
		{"$match": bson.M{"userID": bson.ObjectIdHex(userID), "tool": "scratch"}},
		{"$group": bson.M{"_id": nil, "total": bson.M{"$sum": 1}}},
		{"$project": bson.M{"_id": 0}},
	}
	var total TotalBody
	f := func(col *mgo.Collection) error {
		return col.Pipe(pipeline).One(&total)
	}
	return total, workMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "works", f)

}

// NewScratch 初始化新的scratch作品信息
func (workMod *WorkModels) NewScratch(userID string, name string, tool string, types string, picture string, description string, toolurl string, toolCategory string) *WorkBody {

	nowTime := time.Now().Unix()
	id := bson.NewObjectId()
	fileName := id.Hex() + ".sb3"
	relpath := path.Join("asset", "works", fileName)

	work := &WorkBody{
		BaseBody: BaseBody{
			ID:          id,
			UserID:      bson.ObjectIdHex(userID),
			Name:        name,
			Tool:        tool,
			Picture:     picture,
			Description: description,
			Types:       types,
		},
		Relpath:    relpath,
		ToolURL:    toolurl,
		CreateTime: nowTime,
		Data:       "",
		Category:   toolCategory,
	}
	return work
}

func (workMod *WorkModels) GetProjectByID(userID string, paging PagingInfo) ([]WorkListBody, error) {

	offset := paging.Offset()
	limit := paging.Limit()

	var works []WorkListBody

	fmt.Println(userID)
	fmt.Println(offset)
	fmt.Println(limit)

	f := func(col *mgo.Collection) error {
		return col.Find(bson.M{"userID": bson.ObjectIdHex(userID), "tool": "scratch"}).Select(bson.M{"_id": 1, "name": 1, "picture": 1, "description": 1, "createTime": 1, "data": 1, "tool": 1, "public": 1}).Limit(limit).Skip(offset).All(&works)
	}

	err := workMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "works", f)

	return works, err

}

func (workMod *WorkModels) GetClassStudents(classCode, contentID string) (interface{}, error) {
	result := []bson.M{}

	pipeline := []bson.M{
		{"$match": bson.M{"code": classCode}},
		{"$unwind": "$students"},
		{"$lookup": bson.M{
			"from": "works",
			"let":  bson.M{"userID": "$students.userID"},
			"pipeline": []bson.M{
				{"$match": bson.M{"contentID": bson.ObjectIdHex(contentID)}},
				{"$match": bson.M{"$expr": bson.M{"$eq": []interface{}{"$userID", "$$userID"}}}},
				//{"$match": bson.M{"$and": []interface{}{bson.M{"contentID": bson.ObjectIdHex(contentID)}, bson.M{"$expr": bson.M{"$eq": []interface{}{"$userID", "$$userID"}}}}}},
				{"$lookup": bson.M{
					"from": "users",
					"let":  bson.M{"id": "$userID"},
					"pipeline": []bson.M{
						{"$match": bson.M{"$expr": bson.M{"$eq": []interface{}{"$_id", "$$id"}}}},
						{"$project": bson.M{"_id": 0, "realname": 1}},
					},
					"as": "user",
				}},
			},
			"as": "work",
		},
		},
		{"$unwind": "$work"},
		{"$project": bson.M{
			"_id": 0,
			"id":  "$_id",
			//"work":             1,
			"work.id":          "$work._id",
			"work.browse":      1,
			"work.category":    1,
			"work.contentID":   1,
			"work.createTime":  1,
			"work.data":        1,
			"work.description": 1,
			"work.edit":        1,
			"work.favor":       1,
			"work.laud":        1,
			"work.name":        1,
			"work.originID":    1,
			"work.picture":     1,
			"work.public":      1,
			"work.relpath":     1,
			"work.tool":        1,
			"work.toolURL":     1,
			"work.types":       1,
			"work.user":        1,
			"work.userID":      1,
		}},
	}
	f := func(col *mgo.Collection) error {
		return col.Pipe(pipeline).All(&result)
	}
	err := workMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "classes", f)
	return &result, err
}
