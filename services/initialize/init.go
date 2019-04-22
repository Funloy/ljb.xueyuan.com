package initialize

import (
	"encoding/json"
	"io/ioutil"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"golang.org/x/crypto/bcrypt"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	m "maiyajia.com/models"
	"maiyajia.com/services/daemon"
	"maiyajia.com/services/mongo"
)

// Syncdb 同步数据库数据
func Syncdb() error {

	// 获取数据库
	dbclient := &mongo.MgoClient{}
	dbclient.StartSession()
	defer dbclient.CloseSession()

	if err := initAdminAccount(dbclient); err != nil {
		//return errors.New("创建管理员账号失败，请检查数据库连接是否正常。")
		return err
	}
	if err := initSystemAccount(dbclient); err != nil {
		//return errors.New("创建产品账号失败，请检查产品配置文件或数据库连接是否正常。")
		return err
	}
	return nil
}

func initSystemAccount(dbclient *mongo.MgoClient) error {
	bytes, err := ioutil.ReadFile("./conf/product.json")
	if err != nil {
		beego.Error(err)
	}
	var account *daemon.Account
	// 读取配置文件关于系统账号的信息
	if err := json.Unmarshal(bytes, &account); err != nil {
		return err
	}
	logs.Info("account:", account)
	// 配置系统账号的其他信息
	account.Metadata = "account"
	account.CreateTime = time.Now()
	f := func(col *mgo.Collection) error {
		query := bson.M{"metadata": "account"}
		change := mgo.Change{
			//Update:    bson.M{"$set": bson.M{"name": account.Name, "key": account.Key, "serial": account.Serial, "os": account.OS, "version": account.Version, "newver": account.Newver, "upgrade": account.Upgrade}},
			Update:    account,
			ReturnNew: true,
			Upsert:    true,
		}
		_, err := col.Find(query).Apply(change, nil)
		return err
	}
	return dbclient.Do(beego.AppConfig.String("MongoDB"), "system", f)

}

func initAdminAccount(dbclient *mongo.MgoClient) error {
	// 创建管理员账号
	username := "admin"
	password, _ := hashPassword("admin")
	realname := "管理员"
	gender := 1
	admin := m.NewTeacher(username, password, realname, gender)
	admin.Role.Name = m.ROLE_ADMIN
	count, _ := QueryAdminAccount(username, dbclient)
	if count > 0 {
		return nil
	}
	f := func(col *mgo.Collection) error {
		return col.Insert(admin)
	}
	return dbclient.Do(beego.AppConfig.String("MongoDB"), "users", f)
}

//密码加密
func hashPassword(password string) ([]byte, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	return hash, nil
}

// QueryAdminAccount 查寻管理员用户
func QueryAdminAccount(username string, dbclient *mongo.MgoClient) (int, error) {
	query := bson.M{"username": username}
	var count int
	var err error
	f := func(col *mgo.Collection) error {
		count, err = col.Find(query).Count()
		return err
	}

	return count, dbclient.Do(beego.AppConfig.String("MongoDB"), "users", f)
}
