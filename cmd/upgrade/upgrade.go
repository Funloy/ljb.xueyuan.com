// @Title 升级命令行工具
// @Description 升级命令行工具，当系统下载完成升级文件后，这个工具能热启动服务器。
// @Contact xuchuangxin@icanmake.cn
// @Date 2018-04-22
package main

import (
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/astaxie/beego/logs"
)

func main() {
	//获取进程号
	logs.Info("os.Args[1]:", os.Args[1])
	pid, err := readPid()
	if err != nil {
		panic(err)
	}
	//终止进程
	err = killProc(pid)
	if err != nil {
		panic(err)
	}
	logs.Info("进程终止")
	// 根据不同操作系统，启动服务器
	// switch goos := runtime.GOOS; goos {
	// case "linux":
	err = executeFile(os.Args[1])
	if err != nil {
		logs.Info("reload fail")
	}
	//}

}

func readPid() (int, error) {
	dat, err := ioutil.ReadFile("pid.tmp")

	if err != nil {
		panic(err)
	}
	line := strings.Split(string(dat), "\n")
	return strconv.Atoi(line[0])
}
func killProc(pid int) error {
	proc, err := os.FindProcess(pid)
	if err != nil {
		logs.Error(err)
		return err
	}
	err = proc.Kill()
	if err != nil {
		panic(err)
		return err
	}
	return nil
}

// func reload(dstName string) error {
// 	goos := runtime.GOOS
// 	var dstName string
// 	switch goos {
// 	case "windows":
// 		dstName = beego.AppConfig.String("upgrade_filename")
// 	case "linux":
// 		dstName = beego.AppConfig.String("upgrade_linux_filename")
// 	}
// 	executeFile(dstName)
// 	return nil
// }

//运行文件
func executeFile(dstName string) error {
	cmd := exec.Command(dstName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	if err != nil {
		logs.Error("cmd err", err)
	}
	return err
}
