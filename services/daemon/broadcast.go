// @APIVersion 1.0.0
// @Title UDP广播服务
// @Description 提供UDP广播服务，为客户端发现服务端的IP地址提供服务
// @Author xuchuangxin@icanmake.cn
// @Date 2018-04-26

package daemon

import (
	"github.com/astaxie/beego/logs"
	"maiyajia.com/services/daemon/discovery"
)

// Broadcast 服务器发起局域网广播，接受客户端查询服务端的IP地址
func Broadcast() {
	logs.Info("启动服务器广播")
	go func() {
		discovery.Discover(discovery.Settings{
			Limit: -1,
			//TimeLimit: 60 * 60 * 24 * time.Second,
			Payload: []byte("maiyajia.local"),
		})
		logs.Info("结束广播服务")
	}()

}
