package main

import (
	"fmt"
	"os"
	"time"

	"maiyajia.com/controllers"
	_ "maiyajia.com/routers"
	"maiyajia.com/services/daemon"
	"maiyajia.com/services/initialize"
	"maiyajia.com/services/mongo"
	"maiyajia.com/util"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/plugins/cors"
)

func main() {

	// 开发环境配置项
	if beego.BConfig.RunMode == beego.DEV {
		// beego.BConfig.WebConfig.DirectoryIndex = true
		// beego.BConfig.WebConfig.StaticDir["/swagger"] = "swagger"
	}

	// 生产环境配置项
	if beego.BConfig.RunMode == beego.PROD {
		//设置应用日志
		beego.SetLogger(logs.AdapterFile, `{"filename":"app.log"}`)
		beego.SetLevel(beego.LevelInformational)
	}

	// 设置跨域访问
	beego.InsertFilter("*", beego.BeforeRouter, cors.Allow(&cors.Options{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Authorization", "Access-Control-Allow-Origin", "Access-Control-Allow-Headers", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length", "Access-Control-Allow-Origin", "Access-Control-Allow-Headers", "Content-Type"},
		AllowCredentials: true,
	}))

	// 设置跨域访问
	beego.InsertFilter("/asset/*", beego.BeforeStatic, cors.Allow(&cors.Options{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Access-Control-Allow-Origin", "Access-Control-Allow-Headers", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length", "Access-Control-Allow-Origin", "Access-Control-Allow-Headers", "Content-Type"},
		AllowCredentials: true,
	}))

	// 注册错误处理函数
	beego.ErrorController(&controllers.ErrorController{})

	// 静态文件目录，主要用于头像的存取（后面可以移到专门的静态资源服务器）
	beego.SetStaticPath("/asset", "asset")

	initSystem() // 系统初始化
	go func() {
		for {
			util.RemoveDownFile("tmp")
			time.Sleep(3600 * 24 * 7)
		}
	}()
	beego.Run()
}

// 系统初始化
func initSystem() {
	var isBroadcast bool
	isBroadcast, _ = beego.AppConfig.Bool("isBroadcast")
	//启动mongodb数据库，初始化mongo数据库连接会话
	if err := mongo.Startup(); err != nil {
		beego.Error(err)
		os.Exit(0)
	}
	beego.Informational("Mongodb startup successfully")
	//系统安装
	flag := install()
	// 临时记录程序运行的进程ID
	if flag {
		logPid(os.Getpid())
	}

	// 服务器广播服务
	if isBroadcast {
		daemon.Broadcast()
	}
}

// 系统安装
func install() bool {
	var flag = true
	args := os.Args
	for _, v := range args {
		if v == "-install" {
			flag = false
			beego.Informational("Starting install system...")
			if err := initialize.Syncdb(); err != nil {
				beego.Error("Install system error: ", err)
			}
			installCourses()
			installTools()
			beego.Informational("System install complete! Please reboot system.")
			os.Exit(0)
		}
	}
	logs.Info("flag:", flag)
	return flag

}

// 记录下主线程的PID，用于升级
func logPid(pid int) {
	f, err := os.Create("pid.tmp")
	if err != nil {
		beego.Error("Can not create pid.tmp fie : ", err)
	}
	defer f.Close()

	_, err = f.WriteString(fmt.Sprintf("%d\n", pid))
	if err != nil {
		beego.Error("Can not write pid.tmp : ", err)
	}
}
func installCourses() {
	logs.Info("begin install courses")
	courseCtrl := controllers.CouresController{}
	courseCtrl.Prepare()
	courseCtrl.CourseMod.MgoSession = &courseCtrl.MgoClient
	courseCtrl.InstallCourse()
	courseCtrl.Finish()
	logs.Info("install courses complete")
}
func installTools() {
	logs.Info("begin install tools")
	toolsCtrl := controllers.ToolsController{}
	toolsCtrl.Prepare()
	toolsCtrl.ToolMod.MgoSession = &toolsCtrl.MgoClient
	toolsCtrl.InstallTools()
	toolsCtrl.Finish()
	logs.Info("install tools complete")
}
