// @APIVersion 1.0.0
// @Title 用户模型
// @Description 用户数据模型和用户数据操作方法
// @Contact xuchuangxin@icanmake.cn
// @TermsOfServiceUrl https://maiyajia.com/
// @License
// @LicenseUrl

package models

import (
	"bytes"
	"fmt"
	"html/template"
	"strconv"
	"time"

	"github.com/astaxie/beego"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"maiyajia.com/services/avatar"
	"maiyajia.com/services/mongo"
	"maiyajia.com/services/token"
	"maiyajia.com/util"
)

type UserModels struct {
	MgoSession mongo.MgoClient
}

// UserRecord 所有用户活跃
type UserRecord struct {
	ID          bson.ObjectId `bson:"_id" json:"id"`
	UserID      bson.ObjectId `bson:"userID" json:"userID"`
	UserName    string        `bson:"username" json:"username"`
	LoginClient string        `bson:"loginClient" json:"loginClient"`
	Action      string        `bson:"action" json:"action"`
	CreateTime  int64         `bson:"createTime" json:"createTime"` //创建时间
}

// User 所有用户的基本属性
type User struct {
	ID         bson.ObjectId `bson:"_id" json:"id"`
	Username   string        `bson:"username" json:"username"`
	Password   []byte        `bson:"password" json:"-"`
	Avatar     string        `bson:"avatar" json:"avatar"`
	Realname   string        `bson:"realname" json:"realname"`
	Gender     string        `bson:"gender" json:"gender"`
	Role       *Role         `bson:"role" json:"role"`
	Calculus   int           `bson:"calculus" json:"calculus"`
	Medals     []*Medal      `bson:"medals" json:"medals"`
	Locked     bool          `bson:"locked" json:"locked"`         // 账户是否锁住
	CreateTime time.Time     `bson:"createTime" json:"createTime"` //创建时间
	Email      string        `bson:"email" json:"email"`
}

// Student 学生用户
type Student struct {
	//bson的inline标签，让User的字段跟Student的字段处于同一层级。具体参见https://godoc.org/gopkg.in/mgo.v2/bson
	User        `bson:",inline"`
	Grade       string       `bson:"grade" json:"grade"`
	JoinClasses []*JoinClass `bson:"joinClasses" json:"joinClasses"`
}

// Teacher 教师用户
type Teacher struct {
	//bson的inline标签，参见https://godoc.org/gopkg.in/mgo.v2/bson
	User `bson:",inline"`
}

// JoinClass 用户加入的班级
type JoinClass struct {
	ClassCode string    `bson:"classCode" json:"classCode"`
	JoinTime  time.Time `bson:"joinTime" json:"joinTime"`
}

//QualifyApplication 资格申请
type QualifyApplication struct {
	ID         bson.ObjectId `bson:"_id" json:"id"`
	UserID     bson.ObjectId `bson:"userID" json:"userID"`         //用户ID
	Username   string        `bson:"username" json:"username"`     //用户名
	Name       string        `bson:"name" json:"name"`             //真实姓名
	Mobile     string        `bson:"mobile" json:"mobile"`         //手机号码
	Organize   string        `bson:"organize" json:"organize"`     //学校或机构
	Reason     string        `bson:"reason" json:"reason"`         //申请理由
	Status     int           `bson:"status" json:"status"`         //申请状态；0：拒绝，1：同意，2：审核中
	CreateTime time.Time     `bson:"createTime" json:"createTime"` //申请时间
}

//Apply 用户申请内容
type Apply struct {
	Name     string `json:"name"`     //真实姓名
	Mobile   string `json:"mobile"`   //手机号码
	Organize string `json:"organize"` //学校或机构
	Reason   string `json:"reason"`   //申请理由
}

type Review struct {
	UserID string `json:"userid"` //申请人用户id
	Status int    `json:"status"` //审核状态 0：拒绝，1：同意，2：初始状态
}

//UpdateStatusRole 修改申请状态
func (userMod *UserModels) UpdateStatusRole(userid string) (interface{}, error) {
	result := []bson.M{}
	pipeline := []bson.M{
		{"$match": bson.M{"userID": bson.ObjectIdHex(userid)}},
		{"$lookup": bson.M{
			"from":         "users",
			"localField":   "userID",
			"foreignField": "_id",
			"as":           "inventory_docs",
		}},
		{"$set": bson.M{"user.role.name": "teacher"}},
	}

	f := func(col *mgo.Collection) error {
		return col.Pipe(pipeline).All(&result)
	}
	return &result, userMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "application", f)
}

//UpdateUserRole 修改申请角色
func (userMod *UserModels) UpdateUserRole(userid string) error {
	update := bson.M{"$set": bson.M{"role.name": "teacher"}}
	f := func(col *mgo.Collection) error {
		return col.Update(bson.M{"_id": bson.ObjectIdHex(userid)}, update)
	}
	return userMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "users", f)
}

//UpdateAplStatus 修改申请状态
func (userMod *UserModels) UpdateAplStatus(status int, userid string) error {
	update := bson.M{"$set": bson.M{"status": status, "createTime": time.Now()}}
	f := func(col *mgo.Collection) error {
		return col.Update(bson.M{"userID": bson.ObjectIdHex(userid)}, update)
	}
	return userMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "application", f)
}

//QueryApplicationList 管理员查询所有申请
func (userMod *UserModels) QueryApplicationList() ([]QualifyApplication, error) {
	var qualifylist []QualifyApplication
	f := func(col *mgo.Collection) error {
		return col.Find(bson.M{"status": 2}).All(&qualifylist)
	}
	err := userMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "application", f)
	return qualifylist, err
}

//QueryApplication 查询申请
func (userMod *UserModels) QueryApplication(userid string) ([]QualifyApplication, error) {
	var qualify []QualifyApplication
	f := func(col *mgo.Collection) error {
		return col.Find(bson.M{"userID": bson.ObjectIdHex(userid)}).All(&qualify)
	}
	err := userMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "application", f)
	return qualify, err
}

//NewApplication 初始化资格申请信息
func (userMod *UserModels) NewApplication(userid, username string, apply Apply) error {
	qualify := QualifyApplication{
		ID:         bson.NewObjectId(),
		UserID:     bson.ObjectIdHex(userid),
		Username:   username,
		Name:       apply.Name,
		Mobile:     apply.Mobile,
		Organize:   apply.Organize,
		Reason:     apply.Reason,
		Status:     2,
		CreateTime: time.Now(),
	}
	f := func(col *mgo.Collection) error {
		return col.Insert(qualify)
	}
	return userMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "application", f)
}

// NewStudent 初始化新的学生账户
func NewStudent(username string, password []byte, realname string, gender int, grade int, class int) *Student {
	student := &Student{
		User: User{
			ID:         bson.NewObjectId(),
			Username:   username,
			Password:   password,
			Avatar:     avatarBuilder(gender, username), //随机头像，把用户名当成随机种子来产生头像
			Realname:   realname,
			Gender:     genderTagBuidler(gender),
			Role:       &Role{Name: ROLE_STUDENT},
			Calculus:   new(RegisterCalculus).Calculus(), // 注册新用户获取积分
			Locked:     false,
			CreateTime: time.Now(),
		},
		Grade: gradeTagBuidler(grade, class),
	}
	return student
}

// NewTeacher 初始化新的教师账户
func NewTeacher(username string, password []byte, realname string, gender int) *Teacher {

	teacher := &Teacher{
		User: User{
			ID:         bson.NewObjectId(),
			Username:   username,
			Password:   password,
			Avatar:     avatarBuilder(gender, username),
			Realname:   realname,
			Gender:     genderTagBuidler(gender),
			Role:       &Role{Name: ROLE_TEACHER},
			Calculus:   new(RegisterCalculus).Calculus(),
			Locked:     false,
			CreateTime: time.Now(),
		},
	}
	return teacher
}

//SendRetrieveEmail 发送找回密码邮件
func SendRetrieveEmail(activateToken, email string) error {
	// 发送邮件参数
	from_mail := beego.AppConfig.String("from_mail")
	smtp_server := "smtp.exmail.qq.com"
	smtp_port := 587
	smtp_user := beego.AppConfig.String("email_user")
	smtp_password := beego.AppConfig.String("email_password")
	retrieveLink := beego.AppConfig.String("resetpass_api")
	content := fmt.Sprintf(retrieveLink, activateToken)
	smtpServer := util.NewSmtpSever(smtp_server, smtp_port, smtp_user, smtp_password)
	tpl, _ := template.ParseFiles("conf/retrieve_email.tpl")
	var body bytes.Buffer
	tpl.Execute(&body, struct {
		Message string
	}{
		Message: content,
	})
	sendMail := util.NewSendMail(from_mail, smtp_user, []string{email}, nil, "[麦芽+学习平台]重置您的密码", string(body.Bytes()))
	err := util.SendSmtp(smtpServer, sendMail)
	if err != nil {
		beego.Error("SendRetrieveEmail err: ", err)
		return err
	}
	return nil
}

//UpsertEmail 写入或更新邮箱绑定
func (userMod *UserModels) UpsertEmail(activetoken *token.EmailToken) error {
	upsertfun := func(col *mgo.Collection) error {
		_, err := col.Upsert(
			bson.M{"_id": bson.ObjectIdHex(activetoken.UserID)},
			bson.M{
				"$set": bson.M{
					"email": activetoken.Email,
				},
			})
		return err
	}
	return userMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "users", upsertfun)
}

//SendActivationEmail 发送激活邮件
func SendActivationEmail(activateToken, email, username string) error {
	// 发送邮件参数
	from_mail := beego.AppConfig.String("from_mail")
	smtp_server := "smtp.exmail.qq.com"
	smtp_port := 587
	smtp_user := beego.AppConfig.String("email_user")
	smtp_password := beego.AppConfig.String("email_password")
	activateLink := beego.AppConfig.String("activate_api")
	content := fmt.Sprintf(activateLink, activateToken)
	smtpServer := util.NewSmtpSever(smtp_server, smtp_port, smtp_user, smtp_password)
	tpl, _ := template.ParseFiles("conf/identification_email.tpl")
	var body bytes.Buffer
	tpl.Execute(&body, struct {
		UserName string
		Message  string
	}{
		UserName: username,
		Message:  content,
	})
	sendMail := util.NewSendMail(from_mail, smtp_user, []string{email}, nil, "[麦芽+学习平台]邮箱绑定", string(body.Bytes()))
	err := util.SendSmtp(smtpServer, sendMail)
	if err != nil {
		beego.Error("SendActivationCode err: ", err)
		return err
	}
	return nil
}

// InsertUserRecord 记录用户登陆记录
func (userMod *UserModels) InsertUserRecord(user *User) error {
	f := func(col *mgo.Collection) error {
		query := bson.M{"userID": user.ID}
		change := mgo.Change{
			Update:    bson.M{"userID": user.ID, "username": user.Username, "createTime": time.Now().Unix()},
			ReturnNew: true,
			Upsert:    true,
		}
		_, err := col.Find(query).Apply(change, nil)
		return err
		// return col.Insert(userrecord)
	}
	return userMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "userrecord", f)
}

// RegisteredUser 在数据库创建一个新的用户
func (userMod *UserModels) RegisteredUser(user interface{}) error {
	f := func(col *mgo.Collection) error {
		return col.Insert(user)
	}
	return userMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "users", f)
}

// LockedUser 冻结或解冻用户账号
func (userMod *UserModels) LockedUser(username string, locked bool) error {
	f := func(col *mgo.Collection) error {
		return col.Update(bson.M{"username": username}, bson.M{"$set": bson.M{"locked": locked}})
	}
	return userMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "users", f)
}

// FindUser 根据健值查询用户
func (userMod *UserModels) FindUser(query interface{}) (*User, error) {
	var user *User
	f := func(col *mgo.Collection) error {
		return col.Find(query).One(&user)
	}
	err := userMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "users", f)
	return user, err
}

// FindUserByEmail 通过用户名查找用户,返回用户的基本信息
func (userMod *UserModels) FindUserByEmail(email string) (*User, error) {
	query := bson.M{"email": email}
	return userMod.FindUser(query)
}

// FindUserByUsername 通过用户名查找用户,返回用户的基本信息
func (userMod *UserModels) FindUserByUsername(username string) (*User, error) {
	query := bson.M{"username": username}
	return userMod.FindUser(query)
}

// FindUserByID 通过用户ID查找用户,返回用户的基本信息
func (userMod *UserModels) FindUserByID(uid bson.ObjectId) (*User, error) {
	query := bson.M{"_id": uid}
	return userMod.FindUser(query)
}

// FetchGivenUserInfo 根据用户ID获取用户的指定信息
// fields 参数就是需要取回的用户信息字段
func (userMod *UserModels) FetchGivenUserInfo(uid bson.ObjectId, fields ...string) (*User, error) {
	selected := make(map[string]int)
	for _, v := range fields {
		selected[v] = 1
	}
	var user *User
	f := func(col *mgo.Collection) error {
		return col.FindId(uid).Select(selected).One(&user)
	}
	err := userMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "users", f)
	return user, err
}

// FetchUserProfile 获取用户的详细资料
func (userMod *UserModels) FetchUserProfile(userid bson.ObjectId) (*User, error) {
	var user *User
	f := func(col *mgo.Collection) error {
		return col.Find(bson.M{"_id": userid}).Select(bson.M{"password": 0}).One(&user)
	}
	err := userMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "users", f)
	return user, err
}

// UpdateUser 更新用户信息
func (userMod *UserModels) UpdateUser(username string, update interface{}) error {
	f := func(col *mgo.Collection) error {
		return col.Update(bson.M{"username": username}, update)
	}
	return userMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "users", f)
}

// ResetPassword 重置密码
func (userMod *UserModels) ResetPassword(UserID string, pass []byte) error {
	update := bson.M{"$set": bson.M{"password": pass}}
	f := func(col *mgo.Collection) error {
		return col.Update(bson.M{"_id": bson.ObjectIdHex(UserID)}, update)
	}
	return userMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "users", f)
}

// UpdatePassword 修改通行证密码
func (userMod *UserModels) UpdatePassword(username string, pass []byte) error {
	update := bson.M{"$set": bson.M{"password": pass}}
	return userMod.UpdateUser(username, update)
}

// UpdateAvatar 修改用户头像，返回新头像的URL地址
func (userMod *UserModels) UpdateAvatar(username string, gender int, charset string) (string, error) {
	avatar := avatarBuilder(gender, charset) // 根据性别和种子获取头像
	update := bson.M{"$set": bson.M{"avatar": avatar}}
	return avatar, userMod.UpdateUser(username, update)
}

// IsJoinedClass 检查用户是不是某个班级的成员
func (userMod *UserModels) IsJoinedClass(username, classCode string) bool {
	query := bson.M{"username": username, "joinClasses": bson.M{"$elemMatch": bson.M{"classCode": classCode}}}
	f := func(col *mgo.Collection) error {
		return col.Find(query).One(nil)
	}
	if err := userMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "users", f); err == mgo.ErrNotFound {
		return false
	}
	return true
}

// FetchJoinClasses 查询用户加入的班级详情
func (userMod *UserModels) FetchJoinClasses(username string) (interface{}, error) {

	joinClass := bson.M{}

	// 查询学生加入的班级
	pipeline := []bson.M{
		{"$match": bson.M{"username": username}},
		{"$lookup": bson.M{
			"from":         "classes",
			"foreignField": "code",
			"localField":   "joinClasses.classCode",
			"as":           "classesList",
		},
		},
		{"$project": bson.M{
			// "username":            1,
			// "avatar":              1,
			// "realname":            1,
			// "gender":              1,
			"classesCounts":    bson.M{"$size": "$joinClasses"}, // 统计班级的数量
			"classesList.code": 1,
			"classesList.name": 1,
			"classesList.logo": 1,
		},
		},
	}

	f := func(col *mgo.Collection) error {
		return col.Pipe(pipeline).One(&joinClass)
	}

	if err := userMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "users", f); err != nil {
		beego.Informational(err)
		return nil, err
	}

	return joinClass, nil
}

// UpdateUserCalculus 更新用户微积分
// 更新微积分的同时，检查用户是否可以荣获积分勋章
func (userMod *UserModels) UpdateUserCalculus(uid bson.ObjectId, calculus CalculusInfo) (int, error) {

	result := bson.M{}
	f := func(col *mgo.Collection) error {
		change := mgo.Change{
			Update:    bson.M{"$inc": bson.M{"calculus": calculus.Calculus()}},
			ReturnNew: true,
		}
		_, err := col.FindId(uid).Select(bson.M{"calculus": 1}).Apply(change, &result)
		return err
	}

	if err := userMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "users", f); err != nil {
		return 0, err
	}

	// 检查用户是否可以荣获积分勋章
	totalCalculus, _ := result["calculus"].(int) //类型转换
	medalInfo := new(CalculusMedal)
	medalInfo.SetCalculus(totalCalculus)
	userMod.GrantUserMedal(uid, medalInfo)

	return totalCalculus, nil

}

// GrantUserMedal 授予用户勋章
func (userMod *UserModels) GrantUserMedal(uid bson.ObjectId, medalInfo MedalInfo) *Medal {
	// 检查是否符合勋章的资格，如果符合则返回相应的勋章，反之返回nil
	medal := medalInfo.GetMedal()
	if medal == nil {
		return nil
	}

	exist := false // 判断用户是否已经拿过该勋章
	f := func(col *mgo.Collection) error {
		info, err := col.UpsertId(uid, bson.M{"$addToSet": bson.M{"medals": medal}})
		if info.Updated == 0 {
			exist = true
		}
		return err
	}

	err := userMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "users", f)

	if err != nil || exist {
		return nil
	}

	return medal
}

type totalBody struct {
	Total int `bson:"total" json:"total"`
}

//GetLivenessCount 获取活跃度总数
func (userMod *UserModels) GetLivenessCount(start, end int64) (ActiveCount, error) {
	var total ActiveCount
	pipeline := []bson.M{
		{"$match": bson.M{"createTime": bson.M{"$gt": start, "$lt": end}}},
		{"$group": bson.M{"_id": nil, "total": bson.M{"$sum": 1}}},
		{"$project": bson.M{"_id": 0}},
	}
	f := func(col *mgo.Collection) error {
		return col.Pipe(pipeline).One(&total)
	}
	return total, userMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "userrecord", f)
}

//GetUsersCount 用户总数
func (userMod *UserModels) GetUsersCount() (int, error) {
	var count int
	var err error
	f := func(col *mgo.Collection) error {
		count, err = col.Find(nil).Count()
		return err
	}
	return count, userMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "users", f)
}

//UpsertUserLive 更新当月用户活跃数
func (userMod *UserModels) UpsertUserLive(liveness ActiveCount) error {
	f := func(col *mgo.Collection) error {
		query := bson.M{"time": liveness.Time}
		change := mgo.Change{
			Update:    liveness,
			ReturnNew: true,
			Upsert:    true,
		}
		_, err := col.Find(query).Apply(change, nil)
		return err
	}
	return userMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "liveness", f)

}

//InsertUserLive 用户月活跃统计记录
func (userMod *UserModels) InsertUserLive(liveness []ActiveCount) error {
	var docs []interface{}
	for _, v := range liveness {
		docs = append(docs, v)
	}
	f := func(col *mgo.Collection) error {
		return col.Insert(docs...)
	}
	return userMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "liveness", f)
}

//GetUserLive 获取用户月活跃统计记录
func (userMod *UserModels) GetUserLive(startTime, deadlineTime time.Time) ([]ActiveCount, error) {
	var liveness []ActiveCount
	query := bson.M{"time": bson.M{"$gte": startTime, "$lt": deadlineTime}}
	f := func(col *mgo.Collection) error {
		return col.Find(query).All(&liveness)
	}
	return liveness, userMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "liveness", f)
}
func (userMod *UserModels) DellivenessData() error {
	f := func(col *mgo.Collection) error {
		_, err := col.RemoveAll(bson.M{})
		return err
	}
	userMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "liveness", f)
	return nil
}

/*********************************************************************************************/
/*********************************** 以下为本控制器的内部函数 *********************************/
/*********************************** *********************************************************/

// defaultAvatarBuilder 随机头像
func avatarBuilder(gender int, salt string) string {

	// 性别转换成模块需要的类型
	g := avatar.MALE
	if gender != 0 {
		g = avatar.FEMALE
	}

	// 此处定义头像保存的目录、文件名和图片格式
	// 系统头像的存放规则为： asset/users/avatar/随机文件名.png
	// 当随机生成头像出错时，会返回默认的系统头像，该头像存放在 asset/users/avatar_default.png
	random := strconv.FormatInt(time.Now().UnixNano(), 10)
	avatarDir := "asset/users/avatar/"
	util.CreateDir(avatarDir)
	file := avatarDir + salt + "_" + random + ".png"

	// 根据性别和种子字符串(可选)随机产生头像，并保存到filePath指定的文件
	img, err := avatar.BuilderWithSalt(g, salt)
	if err != nil {
		beego.Error(err)
		return "asset/users/avatar_default.png"
	}

	if err := avatar.SaveAvatarToFile(img, file); err != nil {
		beego.Error(err)
		beego.Debug("SaveAvatarToFile")
		return "asset/users/avatar_default.png"
	}

	return file
}

// gradeTagBuidler 根据年级和班级，生成用户信息中的grade字段信息
func gradeTagBuidler(grade, class int) string {
	var text = "%s%d年%d班"
	switch {
	case grade > 9:
		return fmt.Sprintf(text, "高中", (grade - 9), class)
	case grade > 6:
		return fmt.Sprintf(text, "初中", (grade - 6), class)
	default:
		return fmt.Sprintf(text, "小学", grade, class)
	}
}

func genderTagBuidler(gender int) string {
	switch gender {
	case 0:
		return "男"
	case 1:
		return "女"
	default:
		return "未知"
	}
}
