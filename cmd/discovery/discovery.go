// @APIVersion 1.0.0
// @Title UDP广播服务
// @Description 提供UDP广播服务，为客户端发现服务端的IP地址提供服务
// @Author xuchuangxin@icanmake.cn
// @Date 2018-04-26

package main

import (
	"flag"
	"fmt"
	"os"

	"maiyajia.com/services/daemon/discovery"
)

var (
	h        bool
	duration int
	hostname string
)

func init() {
	flag.BoolVar(&h, "h", false, "帮助")
	flag.IntVar(&duration, "t", 5, "扫描时长(单位秒)")
	flag.StringVar(&hostname, "s", "maiyajia.local", "服务主机名")
	flag.Usage = usage
}

func main() {

	flag.Parse()
	if h {
		flag.Usage()
		os.Exit(0)
	}
	if duration == 0 || len(hostname) == 0 {
		flag.Usage()
		os.Exit(0)
	}

	//timeLimit := time.Duration(duration) * time.Second
	discoveries, err := discovery.Discover(discovery.Settings{
		Limit: -1,
		//TimeLimit: timeLimit,
	})

	if err != nil {
		panic(err)
	} else {
		if len(discoveries) > 0 {
			for _, d := range discoveries {
				if string(d.Payload) == "maiyajia.local" {
					fmt.Printf("%s", d.Address)
					return
				}
			}
		}
	}
}

func usage() {
	fmt.Fprintf(os.Stderr,
		`discovery 版本: discovery/v1.0.0
使用: discovery [-h] [-t 扫描时长(单位秒)] [-s 服务主机名]
参数选项:`)
	fmt.Println()
	flag.PrintDefaults()
}
