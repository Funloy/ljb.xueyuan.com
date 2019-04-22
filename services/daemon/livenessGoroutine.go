package daemon

import (
	"time"

	"github.com/astaxie/beego/logs"
	m "maiyajia.com/models"
	"maiyajia.com/services/mongo"
)

type LivenessModel struct {
	MgoSession *mongo.MgoClient
	UserMod    m.UserModels
}

// UpdateUserLiveness 添加用户活跃度
func (liveness *LivenessModel) UpdateUserLiveness(endyear, endmonth int) error {
	if endyear == time.Now().Year() && endmonth == int(time.Now().Month()) {
		monStartTime := time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.UTC)
		activeCount, err := liveness.UserMod.GetLivenessCount(monStartTime.Unix(), time.Now().Unix())
		if err != nil && "not found" == err.Error() {
			activeCount.Total = 0
		} else if err != nil {
			return err
		}
		activeCount.Time = monStartTime
		if err := liveness.UserMod.UpsertUserLive(activeCount); err != nil {
			logs.Error("UpsertUserLive err:", err)
			return err
		}
	}
	return nil
}
