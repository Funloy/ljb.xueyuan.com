package util

import (
	"bytes"
	srand "crypto/rand"
	"errors"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
)

const dictionary = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"

//CreateRandomString 创建一个随机的字符串, 长度由参数指定
func CreateRandomString(length int) string {
	b := make([]byte, length)
	l := len(dictionary)

	_, err := srand.Read(b)

	if err != nil {
		// 注意: 如果发生错误，则切换到不安全的随机（这种方式产生的随机数不安全）
		rand.Seed(time.Now().UnixNano())
		for i := range b {
			b[i] = dictionary[rand.Int()%l]
		}
	} else {
		for i, v := range b {
			b[i] = dictionary[v%byte(l)]
		}
	}

	return string(b)
}

// CheckFileIsExist 检查文件是否存在
func CheckFileIsExist(file string) error {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return errors.New("File does not exist")
	}
	return nil
}

// CreateProductSerial 创建形如 xxxx-xxxx-xxxx-xxxx形式的序列号
func CreateProductSerial() string {
	var serail []string
	for i := 0; i < 4; i++ {
		serail = append(serail, CreateRandomString(4))
	}
	return strings.Join(serail, "-")
}

//RandomInt 返回[min, max]返回内的随机整数
func RandomInt(min int, max int) int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(max-min) + min
}

//获取登陆网盘 token
func NetdiskAuth() ([]byte, error) {
	loginURL := beego.AppConfig.String("upgrade_login_url")
	loginJSON := beego.AppConfig.String("upgrade_login_json")
	loginData := bytes.NewBuffer([]byte(loginJSON))
	bodyType := "application/json;charset=utf-8"
	tokenResp, err := http.Post(loginURL, bodyType, loginData)
	if err != nil {
		return nil, err
	}
	token, err := ioutil.ReadAll(tokenResp.Body)
	if err != nil {
		return nil, err
	}

	return token, nil
}

//CreateDir 判断文件夹是否存在，不存在则创建
func CreateDir(path string) error {
	_, err := os.Stat(path)
	if err == nil {
		logs.Info(" has dir %v\n", path)
	}
	if os.IsNotExist(err) {
		// 创建文件夹
		error := os.MkdirAll(path, os.ModePerm)
		if error == nil {
			logs.Info("mkdir success!\n")
			return nil
		}
		return error
	}
	return nil
}

//RemoveDownFile 清除下载的文件
func RemoveDownFile(dirfile string) error {
	//当前系统时间
	ct := int32(time.Now().Unix())
	st := int32(3600 * 24 * 7)
	//读取文件目录信息
	dirpath := path.Join(beego.AppPath, dirfile)
	files, _ := ioutil.ReadDir(dirpath)
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		//文件修改时间
		mt := int32(file.ModTime().Unix())
		//清除一天前的文件
		if mt < ct-st {
			os.Remove(path.Join(dirpath, file.Name()))
		}
	}
	return nil
}
