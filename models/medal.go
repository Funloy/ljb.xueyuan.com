// @APIVersion 1.0.0
// @Title 用户勋章模型
// @Description 用户勋章的数据模型和操作方法
// @Contact xuchuangxin@icanmake.cn
// @TermsOfServiceUrl https://maiyajia.com/
// @License
// @LicenseUrl

package models

import (
	"fmt"
	"io/ioutil"

	"github.com/astaxie/beego"
	"github.com/tidwall/gjson"
)

// medalJsonBytes 从文件中读取勋章Json文件后，保存勋章数据的字节数组
var medalJsonBytes []byte

// 初始化动作，读取勋章的配置文件，转换为[]byte字节数组
func init() {
	bytes, err := ioutil.ReadFile(beego.AppConfig.String("medal_path"))
	if err != nil {
		beego.Error(err)
	}
	medalJsonBytes = bytes
}

// Medal 勋章的基本信息
type Medal struct {
	Name        string `bson:"name" json:"name"`
	Image       string `bson:"image" json:"image"`
	Description string `bson:"description" json:"description"`
}

// MedalInfo 勋章接口类型
type MedalInfo interface {
	GetMedal() *Medal // 授予勋章
}

// SignMedal 签到勋章
type SignMedal struct {
	days int
}

// SetDays 设置连续签到的天数
func (sm *SignMedal) SetDays(days int) {
	sm.days = days
}

// GetMedal 授予签到勋章
func (sm *SignMedal) GetMedal() *Medal {

	// 检查签到天数是否达到签到勋章的要求
	query := fmt.Sprintf(`medals.#.entities.#[days="%d"]`, sm.days)
	result := gjson.ParseBytes(medalJsonBytes).Get(query).Array()
	if len(result) == 0 {
		return nil
	}
	// 签到勋章
	medal := &Medal{
		Name:        result[0].Get("name").String(),
		Image:       result[0].Get("image").String(),
		Description: result[0].Get("description").String(),
	}
	return medal
}

// CalculusMedal 积分勋章
type CalculusMedal struct {
	calculus int
}

// SetCalculus 设置用户的微积分
func (cm *CalculusMedal) SetCalculus(calculus int) {
	cm.calculus = calculus
}

// GetMedal 授予积分勋章
func (cm *CalculusMedal) GetMedal() *Medal {
	// 检查积分是否到达积分勋章的要求
	var medal *Medal
	result := gjson.GetBytes(medalJsonBytes, `medals.#.entities`).Array()
	for _, entity := range result {
		for _, item := range entity.Array() {
			min := item.Get("calculus.0")
			max := item.Get("calculus.1")
			if int64(cm.calculus) >= min.Int() && int64(cm.calculus) <= max.Int() {
				medal = &Medal{
					Name:        item.Get("name").String(),
					Image:       item.Get("image").String(),
					Description: item.Get("description").String(),
				}
			}
		}
	}

	return medal
}

// FetchAllMedals 获取全部勋章信息
func FetchAllMedals() interface{} {
	result := gjson.ParseBytes(medalJsonBytes).Value()
	return result
}
