// @APIVersion 1.0.0
// @Title 用户微积分模型
// @Description 用户微积分是系统的用户积分制，名为“微积分”
// @Contact xuchuangxin@icanmake.cn
// @TermsOfServiceUrl https://maiyajia.com/
// @License
// @LicenseUrl

package models

import (
	"github.com/astaxie/beego"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"maiyajia.com/services/mongo"
)

type CalModels struct {
	MgoSession *mongo.MgoClient
}

const (
	// POINTSTAGE 学完一个课时后加积分基数
	POINTSTAGE = 5
)

// CalculusInfo 积分接口
type CalculusInfo interface {
	Calculus() int
}

// FinishedcouCalculus 用户学完课时微积分奖励
type FinishedcouCalculus struct {
}

// Calculus 用户学完微积分的奖励分数
func (fc *FinishedcouCalculus) Calculus() int {
	return POINTSTAGE
}

// RegisterCalculus 注册新用户微积分奖励
type RegisterCalculus struct{}

// Calculus 注册新用户微积分的奖励分数
func (rc *RegisterCalculus) Calculus() int {
	return 100
}

// SignCalculus 每日签到微积分奖励
type SignCalculus struct {
	days int
}

// Calculus 返回每日签到微积分奖励分数
func (sc *SignCalculus) Calculus() int {
	// 算法是：奖励积分 = （(连续签到的天数)2次方） * 20
	return sc.days * sc.days * 20
}

// SetDays 设置连续签到的天数
func (sc *SignCalculus) SetDays(days int) {
	sc.days = days
}

// CalculusRank 查询用户积分排名
func (calMod *CalModels) CalculusRank() (interface{}, error) {
	result := []bson.M{}
	// 查询用户积分并进行排名
	pipeline := []bson.M{
		{"$sort": bson.M{"calculus": -1}},
		{"$project": bson.M{
			"username": 1,
			"avatar":   1,
			"realname": 1,
			"gender":   1,
			"calculus": 1,
		},
		},
		{"$limit": 10},
	}
	f := func(col *mgo.Collection) error {
		return col.Pipe(pipeline).All(&result)
	}
	return &result, calMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "users", f)
}

// CalculusRankInClass 查询用户积分排名
func (calMod *CalModels) CalculusRankInClass(classCode string) (interface{}, error) {
	result := bson.M{}
	// 查询用户积分并进行排名

	pipeline := []bson.M{
		{"$match": bson.M{"code": classCode}},

		{"$lookup": bson.M{
			"from":         "users",
			"localField":   "students.userID",
			"foreignField": "_id",
			"as":           "student",
		},
		},
		{"$project": bson.M{
			"student.username": 1,
			"student.avatar":   1,
			"student.realname": 1,
			"student.gender":   1,
			"student.calculus": 1,
		},
		},

		// {"pipeline": []bson.M{
		// 	{"$sort": bson.M{"students.calculus": -1}},
		// 	{"$limit": 3},
		// },
		// },
	}
	f := func(col *mgo.Collection) error {
		return col.Pipe(pipeline).One(&result)
	}
	err := calMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "classes", f)
	beego.Error(err)
	return &result, err
}
