// @APIVersion 1.0.0
// @Title 后台服务
// @Description 后台服务提供产品与麦芽+远程服务接口的交互
// @Author xuchuangxin@icanmake.cn
// @Date 2018-04-16

package daemon

import (
	"encoding/json"
	"io/ioutil"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"maiyajia.com/services/mongo"
)

type AccountModels struct {
	MgoSession *mongo.MgoClient
}

// Account 产品的账号信息
type Account struct {
	// 账号元数据，当查询数据库时，使用这个字段来读取账号信息。初始化的时候，该字段设置为"account"
	Metadata   string    `bson:"metadata" json:"-"`
	Name       string    `bson:"name" json:"name"`
	Key        string    `bson:"key" json:"key"`
	Serial     string    `bson:"serial" json:"serial"`
	OS         string    `bson:"os" json:"os"`
	Version    string    `bson:"version" json:"version"`
	CreateTime time.Time `bson:"createTime" json:"createTime"`
	Newver     bool      `bson:"newver" json:"-"`        //是否发现新版本
	Upgrade    *Upgrade  `bson:"upgrade" json:"upgrade"` //升级信息
}

// Upgrade 升级服务器返回的升级信息
type Upgrade struct {
	Name      string    `bson:"name" json:"name"`
	Version   string    `bson:"version" json:"version"`
	Asset     *Asset    `bson:"asset" json:"asset"`
	Changelog string    `bson:"changelog" json:"changelog"`
	Date      time.Time `bson:"date" json:"date"`
}

// Asset 产品安装包信息
type Asset struct {
	OS     string `bson:"os" json:"os"`
	Source string `bson:"source" json:"source"`
	Hash   string `bson:"hash" json:"hash"`
}

// FetchSystemAccount 返回产品的信息对象
func FetchSystemAccount() *Account {
	var account *Account
	f := func(col *mgo.Collection) error {
		return col.Find(bson.M{"metadata": "account"}).One(&account)
	}
	err := mongo.Client.Do(beego.AppConfig.String("MongoDB"), "system", f)
	if err != nil {
		logs.Info("FetchSystemAccount err:", err)
	}
	return account
}

// SetSystemUpgradeInfo 设置系统升级信息
func SetSystemUpgradeInfo(info *Upgrade) error {
	f := func(col *mgo.Collection) error {
		return col.Update(bson.M{"metadata": "account"}, bson.M{"$set": bson.M{"newver": true, "upgrade": info}})
	}
	return mongo.Client.Do(beego.AppConfig.String("MongoDB"), "system", f)
}

//SetSystemRebootInfo 系统重启后更新版本信息
func SetSystemRebootInfo(version string) error {
	f := func(col *mgo.Collection) error {
		return col.Update(bson.M{"metadata": "account"}, bson.M{"$set": bson.M{"newver": false, "version": version, "createTime": time.Now()}})
	}
	return mongo.Client.Do(beego.AppConfig.String("MongoDB"), "system", f)
}

//GetProductInfo 获取产品key,serial
func GetProductInfo() (key, serial string, err error) {
	bytes, err := ioutil.ReadFile("./conf/product.json")
	if err != nil {
		beego.Error(err)
	}
	var account *Account
	// 读取配置文件关于系统账号的信息
	if err := json.Unmarshal(bytes, &account); err != nil {
		return "", "", err
	}
	return account.Key, account.Serial, nil
}
