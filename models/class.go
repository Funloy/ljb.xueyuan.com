// @APIVersion 1.0.0
// @Title 班级模型
// @Description 班级模型和数据操作方法
// @Contact xuchuangxin@icanmake.cn
// @TermsOfServiceUrl https://maiyajia.com/
// @License
// @LicenseUrl

package models

import (
	"image"
	"strconv"
	"time"

	"github.com/astaxie/beego"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"maiyajia.com/services/avatar"
	"maiyajia.com/services/mongo"
	"maiyajia.com/util"
)

type ClassModels struct {
	UserMod    UserModels
	MgoSession *mongo.MgoClient
}

//CLASS_CODE_LEN 班级的代码长度
const CLASS_CODE_LEN = 6

// Class 班级类型
type Class struct {
	ID         bson.ObjectId `bson:"_id" json:"id"`
	Name       string        `bson:"name" json:"name"`
	Code       string        `bson:"code" json:"code"`
	Logo       string        `bson:"logo" json:"logo"` // 班徽
	Creator    bson.ObjectId `bson:"creator" json:"creator"`
	CreateTime time.Time     `bson:"createTime" json:"createTime"` //创建时间
	Students   []*Classmate  `bson:"students" json:"students"`
}

// Classmate 班级里的同学
type Classmate struct {
	UserID   bson.ObjectId `bson:"userID" json:"userID"`
	JoinTime time.Time     `bson:"joinTime" json:"joinTime"`
}

// NewClass 初始化一个班级
func (classMod *ClassModels) NewClass(creator bson.ObjectId, name string) *Class {
	// 随机生成班级代码,该代码在班级数据库中不能重复
	code := classMod.classCodeBuilder()
	logo := classMod.classLogoBuilder(creator) //创建班级默认的logo
	var class = &Class{
		ID:         bson.NewObjectId(),
		Name:       name,
		Code:       code,
		Logo:       logo,
		Creator:    creator,
		CreateTime: time.Now(),
	}
	return class
}

// CreateClass 在数据库中创建班级
func (classMod *ClassModels) CreateClass(class interface{}) error {
	f := func(col *mgo.Collection) error {
		return col.Insert(class)
	}
	return classMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "classes", f)
}

// DeleteClass 在数据库中删除班级
func (classMod *ClassModels) DeleteClass(code string) error {
	f := func(col *mgo.Collection) error {
		return col.Remove(bson.M{"code": code})
	}
	return classMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "classes", f)
}

//FindClass 根据健值查找班级
func (classMod *ClassModels) FindClass(query interface{}) (Class, error) {
	var class Class
	f := func(col *mgo.Collection) error {
		return col.Find(query).One(&class)
	}
	err := classMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "classes", f)
	return class, err
}

// FindClassByName 根据班级名称查询班级
func (classMod *ClassModels) FindClassByName(name string) (Class, error) {
	query := bson.M{"name": name}
	return classMod.FindClass(query)
}

// FindClassByCode 根据班级代码查询班级
func (classMod *ClassModels) FindClassByCode(code string) (Class, error) {
	query := bson.M{"code": code}
	return classMod.FindClass(query)
}

// HasClass 查询班级是否存在
func (classMod *ClassModels) HasClass(query interface{}) bool {
	f := func(col *mgo.Collection) error {
		return col.Find(query).One(nil)
	}
	if err := classMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "classes", f); err == mgo.ErrNotFound {
		return false
	}
	return true
}

// HasClassName 查询班级是否已经存在 false表示不存在该名称的班级
func (classMod *ClassModels) HasClassName(name string) bool {
	query := bson.M{"name": name}
	return classMod.HasClass(query)
}

// HasClassCode 查询班级代码是否已经存在 false表示不存在该名称代码
func (classMod *ClassModels) HasClassCode(code string) bool {
	query := bson.M{"code": code}
	return classMod.HasClass(query)
}

// IsClassMember 判断用户是不是班级的成员(包括班级创建者)
func (classMod *ClassModels) IsClassMember(uid bson.ObjectId, classCode string) bool {
	query := bson.M{"code": classCode, "$or": []bson.M{bson.M{"students.userID": uid}, bson.M{"creator": uid}}}

	f := func(col *mgo.Collection) error {
		return col.Find(query).One(&bson.M{})
	}

	err := classMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "classes", f)
	beego.Informational(err)
	if err == mgo.ErrNotFound {
		return false
	}
	return true

}

// @Title 学生加入班级
// @Description 把用户添加到班级的学生名单后，再把班级的信息写入到用户信息库里
// @Param	uid	bson.ObjectId	true "用户ID"
// @Param	code	string	true "班级代码"
// @Success  nil {error}
// @Failure  error {error} 如果发生错误，则表示找不到代码对应的班级
func (classMod *ClassModels) AddStudentToClass(uid bson.ObjectId, code string) error {

	joinTime := time.Now()

	// 把用户插入班级数据库
	var mate = &Classmate{
		UserID:   uid,
		JoinTime: joinTime,
	}
	f := func(col *mgo.Collection) error {
		query := bson.M{"code": code, "students.userID": bson.M{"$ne": uid}}
		change := mgo.Change{
			Update:    bson.M{"$push": bson.M{"students": mate}},
			ReturnNew: true,
		}
		_, err := col.Find(query).Apply(change, nil)
		return err
	}

	if err := classMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "classes", f); err != nil {
		return err
	}

	// 把班级信息插入用户数据库
	var joinClass = &JoinClass{
		ClassCode: code,
		JoinTime:  joinTime,
	}
	ff := func(col *mgo.Collection) error {
		query := bson.M{"_id": uid, "joinClasses.classCode": bson.M{"$ne": code}}
		change := mgo.Change{
			Update:    bson.M{"$push": bson.M{"joinClasses": joinClass}},
			ReturnNew: true,
		}
		_, err := col.Find(query).Apply(change, nil)
		return err
	}

	return classMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "users", ff)

}

// @Title 学生移出班级
// @Description 把用户从班级学生里移出
// @Param	uid	bson.ObjectId	true "用户ID"
// @Param	code	string	true "班级代码"
// @Success  nil {error}
// @Failure  error {error} 如果发生错误，则表示找不到代码对应的班级
func (classMod *ClassModels) RemoveStudentFromClass(uid bson.ObjectId, code string) error {

	f := func(col *mgo.Collection) error {
		query := bson.M{"code": code}
		change := mgo.Change{
			Update:    bson.M{"$pull": bson.M{"students": bson.M{"userID": uid}}},
			ReturnNew: true,
		}
		_, err := col.Find(query).Apply(change, nil)
		return err
	}
	if err := classMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "classes", f); err != nil {
		return err
	}

	ff := func(col *mgo.Collection) error {
		query := bson.M{"_id": uid}
		change := mgo.Change{
			Update:    bson.M{"$pull": bson.M{"joinClasses": bson.M{"classCode": code}}},
			ReturnNew: true,
		}
		_, err := col.Find(query).Apply(change, nil)
		return err
	}

	return classMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "users", ff)

}

// @Title 更新班级名称
// @Description 更改班级的名称，<strong>注意的是要确保新的名称不能重复，本方法未做名称重复的检查</strong>
// @Param	classId		string	true "班级ID"
// @Param	creatorId	string	true "班级创建者ID"
// @Param	newname		string	true "班级新名称"
// @Success  nil {error}
// @Failure  error {error} 错误：（1）表示找不到对应的班级，（2）操作者不是班级的创建者
func (classMod *ClassModels) RenameClassName(creatorID, code, name string) error {

	query := bson.M{"code": code, "creator": bson.ObjectIdHex(creatorID)}

	f := func(col *mgo.Collection) error {
		return col.Update(query, bson.M{"$set": bson.M{"name": name}})
	}

	return classMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "classes", f)
}

// FetchClassDetails 获取班级和班级学生的信息
func (classMod *ClassModels) FetchClassDetails(code string) (interface{}, error) {
	result := bson.M{}
	// 查询班级的学生信息
	pipeline := []bson.M{
		{"$match": bson.M{"code": code}},
		{"$lookup": bson.M{
			"from":         "users",
			"foreignField": "_id",
			"localField":   "students.userID",
			"as":           "students",
		},
		},
		{"$project": bson.M{
			"_id":               0,
			"name":              1,
			"logo":              1,
			"code":              1,
			"creator":           1,
			"creatTime":         1,
			"students._id":      1,
			"students.username": 1,
			"students.avatar":   1,
			"students.realname": 1,
			"students.calculus": 1,
			"students.gender":   1,
		},
		},
	}

	f := func(col *mgo.Collection) error {
		return col.Pipe(pipeline).One(&result)
	}

	if err := classMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "classes", f); err != nil {
		return nil, err
	}

	// 查询班级的创建者信息
	creatorID := result["creator"]
	creator := bson.M{}
	ff := func(col *mgo.Collection) error {
		return col.FindId(creatorID).Select(bson.M{"username": 1, "avatar": 1, "realname": 1}).One(&creator)
	}
	// 创建者信息查询失败，不做处理直接
	classMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "users", ff)
	result["creator"] = creator

	return result, nil
}

// ClassesByCreator 查询用户创建的班级列表
func (classMod *ClassModels) ClassesByCreator(creatorID bson.ObjectId) (interface{}, error) {

	classes := []bson.M{}
	query := bson.M{"creator": creatorID}

	f := func(col *mgo.Collection) error {
		return col.Find(query).Select(bson.M{"students": 0}).All(&classes)
	}

	return &classes, classMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "classes", f)
}

// defaultClassCodeBuilder 随机生成班级代码，并且确保跟系统已存在的班级代码不同
func (classMod *ClassModels) classCodeBuilder() string {
	var ok, code = true, ""
	for ok {
		code = util.CreateRandomString(CLASS_CODE_LEN)
		ok = classMod.HasClassCode(code)
	}
	return code

	// 因为这里是循环查询数据库，有可能导致死循环，如果要更保险的话，可以在此处加一个超时控制
	// var ok, code = true, ""
	// timeout := time.After(time.Second * 10)
	// for ok {
	// 	select {
	// 	case <-timeout:
	// 		ok = false
	// 	default:
	// 		code = util.CreateRandomString(CLASS_CODE_LEN)
	// 		ok = HasClassCode(code)
	// 	}
	// }
	// return code

}
func (classMod *ClassModels) classLogoBuilder(creator bson.ObjectId) string {
	// 获取创建者的头像
	u, _ := classMod.UserMod.FetchGivenUserInfo(creator, "avatar")
	creatorAvatar, _ := avatar.ReadAvatarFromFile(u.Avatar)

	//把创建者的头像设置为组合头像中的主要构图
	var images []image.Image
	images = append(images, creatorAvatar)
	// 随机产生几张图片
	num := util.RandomInt(2, 5)
	for i := 0; i < num; i++ {
		var gender avatar.Gender
		if util.RandomInt(1, 2) == 1 {
			gender = avatar.FEMALE
		} else {
			gender = avatar.MALE
		}
		img, _ := avatar.BuilderWithSalt(gender, util.CreateRandomString(6))
		images = append(images, img)
	}

	image := avatar.CompositeBuilder(images)

	salt := strconv.FormatInt(time.Now().UnixNano(), 10)
	file := "asset/users/avatar/class_" + salt + ".png"

	if err := avatar.SaveAvatarToFile(image, file); err != nil {
		return "asset/users/avatar_default.png"
	}

	return file
}
