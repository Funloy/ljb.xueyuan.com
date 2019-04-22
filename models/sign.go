// @APIVersion 1.0.0
// @Title 用户签到模型
// @Description 用户签到模型和签到操作方法
// @Contact xuchuangxin@icanmake.cn
// @TermsOfServiceUrl https://maiyajia.com/
// @License
// @LicenseUrl

package models

import (
	"fmt"
	"time"

	"github.com/astaxie/beego"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"maiyajia.com/services/mongo"
)

type SignModels struct {
	MgoSession *mongo.MgoClient
}

// SignedLog 用户签到记录
type SignedLog struct {
	UserID           bson.ObjectId `bson:"userID" json:"userID"`
	LatestSigned     time.Time     `bson:"latestSigned" json:"latestSigned"`         //最近的签到日期
	ContinuousSigned int           `bson:"continuousSigned" json:"continuousSigned"` //连续签到的天数
	TotalSigned      int           `bson:"totalSigned" json:"totalSigned"`           // 总的签到天数
	SignedTimes      []time.Time   `bson:"signedTimes" json:"-"`                     //记录时间
}

// @Title 用户签到
// @Description 用户每日签到
// @Param	uid	string	true "用户ID"
// @Success  object, true {*SignedLog, bool} "true表示成功签到，false表示今天已经签到"
// @Failure  object, false {*SignedLog bool}
func (signMod *SignModels) SignedLogger(uid bson.ObjectId) (*SignedLog, bool) {

	//签到日期
	today, _ := time.Parse("2006-01-02", time.Now().Format("2006-01-02"))
	//signedTime := time.Now()
	signed, err := signMod.FetchSigned(uid)

	//  最新签到时间等于今天，则返回false表示今天已经签到
	if err == nil && signed.LatestSigned.Equal(today) {
		return nil, false
	}

	// 如果找不到用户签到记录，则新建一个用户签到记录
	if err == mgo.ErrNotFound {
		newSigned := &SignedLog{
			UserID:           uid,
			LatestSigned:     today,
			ContinuousSigned: 1,
			TotalSigned:      1,
			SignedTimes:      []time.Time{today},
		}

		// 插入一个新的签到记录
		f := func(col *mgo.Collection) error {
			return col.Insert(newSigned)
		}
		signMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "signed", f)
		return newSigned, true
	}

	// 更新用户签到信息
	// 获取用户最近的签到信息
	latestSigned := signed.LatestSigned
	continuousSigned := signed.ContinuousSigned

	// 检查用户否在前一天签到，更新用户连续签到的天数
	if (today.Sub(latestSigned).Hours() / 24) == 1 {
		continuousSigned = continuousSigned + 1
	} else {
		continuousSigned = 1
	}

	f := func(col *mgo.Collection) error {

		query := bson.M{"userID": uid}
		update := bson.M{"$set": bson.M{"latestSigned": today, "continuousSigned": continuousSigned},
			"$inc":  bson.M{"totalSigned": 1},
			"$push": bson.M{"signedTimes": today}}
		change := mgo.Change{
			Update:    update,
			ReturnNew: true,
		}
		_, err := col.Find(query).Select(bson.M{"signedTimes": 0}).Apply(change, &signed)
		return err
	}

	if err := signMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "signed", f); err != nil {
		beego.Error(err)
		return nil, true
	}
	return signed, true

}

// SignedRecords 用户签到记录
func (signMod *SignModels) SignedRecords(uid string, start, end time.Time) (days []string, err error) {

	var logs SignedLog

	f := func(col *mgo.Collection) error {
		return col.Find(bson.M{"userID": bson.ObjectIdHex(uid)}).One(&logs)
	}

	if err := signMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "signed", f); err != nil {
		return nil, err
	}
	fmt.Printf("logs:", logs)
	// 比较日期
	for _, value := range logs.SignedTimes {
		if value.After(start) && value.Before(end) || value.Equal(start) || value.Equal(end) {
			days = append(days, value.Format("2006-01-02"))
		}
	}
	return days, err
}

// // HasSigned 查询用户是否签到
// func HasSigned(uid bson.ObjectId, signedTime time.Time) (*SignedLog, error) {

// 	var signed *SignedLog

// 	f := func(col *mgo.Collection) error {
// 		return col.Find(bson.M{"userID": uid, "latestSigned": bson.M{"$qe": signedTime}}).One(signed)
// 	}
// 	return signed, mongo.Client.Do(beego.AppConfig.String("MongoDB"), "signed", f)

// }

// FetchSigned 查找用户的签到记录
func (signMod *SignModels) FetchSigned(uid bson.ObjectId) (*SignedLog, error) {
	var signed *SignedLog

	f := func(col *mgo.Collection) error {
		return col.Find(bson.M{"userID": uid}).Select(bson.M{"signedTimes": 0}).One(&signed)
	}
	return signed, signMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "signed", f)

}
