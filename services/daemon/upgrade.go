package daemon

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/cavaliercoder/grab"
	"github.com/gorilla/websocket"
	"github.com/mholt/archiver"
	"github.com/parnurzeal/gorequest"
	m "maiyajia.com/models"
	"maiyajia.com/services/mongo"
	"maiyajia.com/services/token"
	"maiyajia.com/util"
)

type UpgradeModels struct {
	MgoSession *mongo.MgoClient
	ToolMod    m.ToolModels
	CourseMod  m.CourseModels
}

var loginToken string

//升级阶段常量
const (
	_             = iota
	DOWNLOAD      //下载阶段
	UNCOMPRESSION //解压阶段
	DOEN          //全部完成
	REBOOT
)

type out struct {
	Code int `json:"code"`
	// Message string    `json:"message"`
	State *progress `json:"state"`
}
type progress struct {
	Stage    int `json:"stage"`
	Progress int `json:"progress"`
}
type upgradeResult struct {
	Code    int      `json:"code"`
	Newver  bool     `json:"newver"`
	Upgrade *Upgrade `json:"upgrade"`
}
type UpgradeInfo struct {
	Name      string    `bson:"name" json:"name"`
	Version   string    `bson:"version" json:"version"`
	Changelog string    `bson:"changelog" json:"changelog"`
	Date      time.Time `bson:"date" json:"date"`
}

// UpgradeCheck 升级检查
func UpgradeCheck() (*Upgrade, *UpgradeInfo, error) {
	// 获取产品账户信息，使用key和serial远程登录升级服务器
	account := FetchSystemAccount()
	key := account.Key
	serial := account.Serial
	os := account.OS
	version := account.Version

	// 获取token
	var err error
	loginToken, err = getLoginToken(key, serial)
	if err != nil {
		return nil, nil, errors.New(m.GetErrorMsgs(m.ERR_UPGRADE_LOGIN_TOKEN))
	}

	// 向远程服务器发送升级检查请求
	var upgradeResult *upgradeResult
	params := fmt.Sprintf(`{"os":"%s","version":"%s"}`, os, version)
	auth := fmt.Sprintf(`Bearer %s`, loginToken)
	request := gorequest.New().Timeout(60 * time.Second) // 超时60秒
	_, _, errs := request.Post(beego.AppConfig.String("update_api")).
		Set("Authorization", auth).
		Send(params).
		EndStruct(&upgradeResult)
	if errs == nil && upgradeResult.Code == 0 {
		if upgradeResult.Upgrade == nil {
			return nil, nil, nil
		}
		SetSystemUpgradeInfo(upgradeResult.Upgrade) //把升级信息写入数据库
		upgradeInfo := &UpgradeInfo{
			Name:      upgradeResult.Upgrade.Name,
			Version:   upgradeResult.Upgrade.Version,
			Changelog: upgradeResult.Upgrade.Changelog,
			Date:      upgradeResult.Upgrade.Date,
		}
		return upgradeResult.Upgrade, upgradeInfo, nil
	}
	logs.Info("errs:::", errs)
	beego.Error(errs)
	return nil, nil, errors.New(m.GetErrorMsgs(m.ERR_UPGRADE_CHECK_FAIL))

}

// LaunchUpgrade 产品升级
// 升级步骤：
// （1）下载升级包
// （2）校验升级包
// （3）解压升级包
// （4）备份文件，拷贝升级文件
func LaunchUpgrade(ws *websocket.Conn) error {
	//把升级信息写入数据库
	upgradeinfo, _, _ := UpgradeCheck()
	SetSystemUpgradeInfo(upgradeinfo)
	var out out
	var fileName string
	util.CreateDir("tmp/")
	timestamp := time.Now().Unix()
	switch goos := runtime.GOOS; goos {
	case "windows":
		fileName = fmt.Sprintf("tmp/%d.zip", timestamp)
	case "linux":
		fileName = fmt.Sprintf("tmp/%d.tgz", timestamp)
	}

	account := FetchSystemAccount()
	if !account.Newver {
		send_msg, err := json.Marshal(WsAbortWithError(m.ERR_UPGRADE_DOEN))
		ws.WriteMessage(websocket.TextMessage, []byte(send_msg))
		if err != nil {
			logs.Error("error:", err)
		}
		return errors.New("已经是最新版本")
	}
	//获取升级包的相关信息
	source := account.Upgrade.Asset.Source
	hash := account.Upgrade.Asset.Hash
	logs.Info("acount:", source, hash)
	beego.Informational("准备升级...")
	//创建文件下载客户端（若在七牛云下载需要在此修改）
	client := grab.NewClient()
	req, _ := grab.NewRequest(fileName, source)
	//判断是否为七牛云公开下载
	isQINIU := beego.AppConfig.DefaultBool("isQINIU", false)
	if !isQINIU {
		token, err := util.NetdiskAuth()
		req.HTTPRequest.Header.Add("Authorization", "Bearer "+string(token))
		if err != nil {
			logs.Error("网盘登陆失败：", err)
			return err
		}
	}

	// 配置checksum并检验升级包
	logs.Info("hash:::", hash)
	sum, err := hex.DecodeString(hash)
	if err != nil {
		logs.Error("hash err:", err)
		return err
	}
	req.SetChecksum(sha256.New(), sum, true)
	// 开始下载升级包
	logs.Info("开始下载升级包 %v...\n", req.URL())
	resp := client.Do(req)
	logs.Info("resp.Filename:", resp.Filename)
	logs.Info("  %v\n", resp.HTTPResponse.Status)
	// 每隔500毫秒打印下载进度
	t := time.NewTicker(50 * time.Millisecond)
	defer t.Stop()

Loop:
	for {
		select {
		case <-t.C:
			out.Code = 0
			out.State = &progress{
				Stage:    DOWNLOAD,
				Progress: int(100 * resp.Progress()),
				// Progress: strconv.FormatFloat(100*resp.Progress(), 'f', 2, 32),
			}
			logs.Info("100*resp.Progress():", 100*resp.Progress(), "resp.Size:", resp.Size)
			send_msg, err := json.Marshal(out)
			if err != nil {
				logs.Error("error:", err)
			}
			ws.WriteMessage(websocket.TextMessage, []byte(send_msg))
			logs.Info("  已下载 %v / %v 字节 ",
				resp.BytesComplete(),
				resp.Size)

		case <-resp.Done:
			// 下载完成
			out.Code = 0
			out.State = &progress{
				Stage:    DOWNLOAD,
				Progress: 100,
			}
			send_msg, err := json.Marshal(out)
			ws.WriteMessage(websocket.TextMessage, []byte(send_msg))
			if err != nil {
				logs.Error("error:", err)
			}
			break Loop
		}
	}
	// 检查错误
	if err := resp.Err(); err != nil {
		logs.Info("downloaderror:", err)
		send_msg, err := json.Marshal(WsAbortWithError(m.ERR_UPGRADE_DOWNLOAD))
		ws.WriteMessage(websocket.TextMessage, []byte(send_msg))
		if err != nil {
			logs.Error("error:", err)
			if err := os.RemoveAll(fileName); err != nil {
				logs.Error("remove dir[%s] fail", path.Join(beego.AppPath, "tmp"))
				return err
			}
		}
		return err
	}
	logs.Info("下载完成，文件已经保存到: %v \n", resp.Filename)
	//备份文件,重命名原exe文件
	dstName := os.Args[0]
	appPath := path.Join(beego.AppPath)
	newName := fmt.Sprintf("%vold", dstName)
	if err := rename(dstName, newName); err != nil {
		// logs.Info("rename file")
	}
	files, _ := ioutil.ReadDir(appPath)
	for _, file := range files {
		if file.Name() == "conf" {
			srcDir := file.Name()
			ndestDir := beego.AppConfig.String("backup_conf")
			backupDir(srcDir, ndestDir)
		} else if file.Name() == "asset" {
			srcDir := "asset/medal"
			ndestDir := beego.AppConfig.String("backup_medal")
			backupDir(srcDir, ndestDir)
		}
	}
	webDir := beego.AppConfig.String("windows_web_dir")
	if err := os.RemoveAll(webDir); err != nil {
		logs.Error("remove dir[%s] fail", path.Join(beego.AppPath, webDir))
	}
	// 解压安装包
	out.Code = 0
	out.State = &progress{
		Stage:    UNCOMPRESSION,
		Progress: 0,
	}
	send_msg, err := json.Marshal(out)
	if err != nil {
		logs.Error("error:", err)
	}
	ws.WriteMessage(websocket.TextMessage, []byte(send_msg))
	if err := ExtractPackage(fileName, appPath); err != nil {
		if err := os.RemoveAll(fileName); err != nil {
			logs.Error("remove dir[%s] fail", path.Join(beego.AppPath, "asset", "courses"))
			return err
		}
		return err
	}
	out.Code = 0
	out.State = &progress{
		Stage:    UNCOMPRESSION,
		Progress: 100,
	}
	send_msg, err = json.Marshal(out)
	ws.WriteMessage(websocket.TextMessage, []byte(send_msg))
	//复制前端文件夹
	//copyWeb()
	out.State = &progress{
		Stage:    DOEN,
		Progress: 100,
	}
	send_msg, err = json.Marshal(out)
	ws.WriteMessage(websocket.TextMessage, []byte(send_msg))
	return nil
}

// SysReboot 系统重启
func SysReboot() error {
	// 安装升级包 运行cmd/upgrade 目录下的升级命令行工具,热启动完成升级
	var targetFile string
	switch goos := runtime.GOOS; goos {
	case "windows":
		targetFile = beego.AppConfig.String("upgrade_server_url")
	case "linux":
		targetFile = beego.AppConfig.String("upgrade_linux_server_url")
	}
	logs.Info("targetFile,os.Args[0]:::", targetFile, os.Args[0])
	err := ExecuteFile(targetFile, os.Args[0])
	if err != nil {
		logs.Info("err:", err)
		return err
	}
	return nil
}

//UpdateUpgradeInfo 更新平台升级信息
func UpdateUpgradeInfo() error {
	account := FetchSystemAccount()
	if account.Upgrade == nil {
		return nil
	}
	if err := SetSystemRebootInfo(account.Upgrade.Version); err != nil {
		logs.Info("UpdateUpgradeInfo err:", err)
		return err
	}
	return nil
}

//CheckTools 工具升级
func (upgradeMod *UpgradeModels) CheckTools() (*m.UpgradeTool, error) {
	tools, err := upgradeMod.ToolMod.GetToolsInfo()
	if err != nil {
		logs.Error("GetToolsInfo fail", err)
		return nil, err
	}
	out := make(map[string]interface{})
	out["tools"] = tools
	send_msg, err := json.Marshal(out)
	params := fmt.Sprintf(string(send_msg))
	var upgradeTools *m.UpgradeTool
	request := gorequest.New().Timeout(60 * time.Second)
	_, _, errs := request.Post(beego.AppConfig.String("upgrade_tools")).Send(params).EndStruct(&upgradeTools)
	if errs == nil {
		return upgradeTools, nil
	}
	err = errors.New("call upgrade tool server fail")
	logs.Error("CheckTools() err:", err)
	return nil, err
}

//CheckCourses 新增课程检测
func (upgradeMod *UpgradeModels) CheckCourses() (*m.UpgradeCourses, error) {
	url := beego.AppConfig.String("course_mall_url")
	logs.Info("url:", url)
	productKey, productSerial, err := GetProductInfo()
	if err != nil {
		logs.Error("GetProductInfo fail", err)
		return nil, err
	}
	courses, err := upgradeMod.CourseMod.GetCoursesInfo()
	if err != nil {
		logs.Error("GetCoursesInfo fail", err)
		return nil, err
	}
	out := make(map[string]interface{})
	out["url"] = url
	out["key"] = productKey
	out["serial"] = productSerial
	out["courses"] = courses
	send_msg, err := json.Marshal(out)
	params := fmt.Sprintf(string(send_msg))
	var upgradeCourses *m.UpgradeCourses
	request := gorequest.New().Timeout(60 * time.Second)
	_, _, errs := request.Post(beego.AppConfig.String("upgrade_courses")).Send(params).EndStruct(&upgradeCourses)
	if errs == nil {
		return upgradeCourses, nil
	}

	return nil, nil
}

/*********************************************************************************************/
/*********************************** 以下为本控制器的内部函数 *********************************/
/*********************************** *********************************************************/

// getLoginToken 获取登录远程服务器的Token
func getLoginToken(key, serial string) (string, error) {
	// loginToken不等于0，则需要验证token的有效性，有效则直接返回原来的token
	if len(loginToken) != 0 {
		_, err := token.ValidatesToken(loginToken)
		if err == nil {
			return loginToken, nil
		}
	}
	//token==0 或者 token无效，需要重新登录获取新的token
	token, err := upgradeLogin(key, serial)
	if len(token) == 0 || err != nil {
		return "", err
	}
	return token, nil
}

func upgradeLogin(key, serial string) (string, error) {

	params := fmt.Sprintf(`{"key":"%s","serial":"%s"}`, key, serial)
	logs.Info(" upgrade params:", params)
	// 登录返回数据定义
	var loginResult struct {
		Code  int    `json:"code"`
		Token string `json:"token"`
	}
	request := gorequest.New().Timeout(60 * time.Second)
	_, _, errs := request.Post(beego.AppConfig.String("login_api")).
		Send(params).
		EndStruct(&loginResult)

	if errs == nil && loginResult.Code == 0 {
		return loginResult.Token, nil
	}

	return "", errors.New("token is empty")

}

// ExtractPackage 解压安装包文件
func ExtractPackage(filename string, dist string) error {
	// if err := archiver.Zip.Open(filename, dist); err != nil {
	// 	return err
	// }
	os := runtime.GOOS
	if os == "windows" {
		if !archiver.Zip.Match(filename) {
			return errors.New("It is not .zip file")
		}
		if err := archiver.Zip.Open(filename, dist); err != nil {
			return err
		}
	} else {
		if !archiver.TarGz.Match(filename) {
			return errors.New("It is not .tar.gz file")
		}
		if err := archiver.TarGz.Open(filename, dist); err != nil {
			return err
		}
	}
	return nil
}

//CopyFile 复制文件
func CopyFile(dstName, srcName string) (written int64, err error) {
	src, err := os.Open(srcName)
	if err != nil {
		return
	}
	defer src.Close()
	dst, err := os.OpenFile(dstName, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return
	}
	defer dst.Close()
	return io.Copy(dst, src)
}

//文件重命名
func rename(oldName, newName string) error {
	err := os.Rename(oldName, newName)
	if err != nil {
		//如果重命名文件失败,则输出错误 file rename Error!
		logs.Info("file rename Error!")
		return err
	}
	return nil
}

//backupDir 备份文件夹
func backupDir(src, dest string) {
	srcpath := path.Join(beego.AppPath)
	switch os := runtime.GOOS; os {
	case "windows":
		srcDir := strings.Replace(path.Join(srcpath, src), "/", "\\", -1)
		destDir := strings.Replace(path.Join(srcpath, dest), "/", "\\", -1)
		copyDir(srcDir, destDir)
	case "linux":
		srcDir := path.Join(srcpath, src)
		destDir := path.Join(srcpath, dest)
		copyDir(srcDir, destDir)
	default:
		logs.Info("os:", runtime.GOOS)
	}
}
func copyDir(src string, dest string) {
	src_original := src
	err := filepath.Walk(src, func(src string, f os.FileInfo, err error) error {
		if f == nil {
			return err
		}
		if f.IsDir() {
			//          copyDir(f.Name(), dest+"/"+f.Name())
		} else {
			dest_new := strings.Replace(src, src_original, dest, -1)
			CopyFiles(src, dest_new)
		}
		return nil
	})
	if err != nil {
		fmt.Printf("filepath.Walk() returned %v\n", err)
	}
}

func CopyFiles(src, dst string) (w int64, err error) {

	srcFile, err := os.Open(src)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer srcFile.Close()
	dst_slices := strings.Split(dst, string(filepath.Separator))
	dst_slices_len := len(dst_slices)
	dest_dir := ""
	for i := 0; i < dst_slices_len-1; i++ {
		dest_dir = dest_dir + dst_slices[i] + string(filepath.Separator)
	}
	b, err := PathExists(dest_dir)
	if b == false {
		err := os.MkdirAll(dest_dir, os.ModePerm) //在当前目录下生成md目录
		if err != nil {
			fmt.Println(err)
		}
	}
	dstFile, err := os.Create(dst)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	defer dstFile.Close()

	return io.Copy(dstFile, srcFile)
}
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// WsAbortWithError 根据错误码获取错误描述信息，然后发送到请求客户端

func WsAbortWithError(code int) m.ErrorResult {
	result := m.ErrorResult{
		Code:    code,
		Message: m.GetErrorMsgs(code),
	}
	return result
}

//ExecuteFile 运行文件
func ExecuteFile(targetFile, arg string) error {
	cmd := exec.Command(targetFile, arg)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	if err != nil {
		logs.Error("cmd err", err)
	}
	return err
}

//copyWeb
func copyWeb() {
	var ndestDir string
	srcDir := beego.AppConfig.String("web_unzip_dir")
	// switch goos := runtime.GOOS; goos {
	// case "windows":
	// 	ndestDir = "..\\" + srcDir
	// case "linux":
	// 	ndestDir = beego.AppConfig.String("website_dir")
	// }
	ndestDir = beego.AppConfig.String("website_dir")
	backupDir(srcDir, ndestDir)
}
