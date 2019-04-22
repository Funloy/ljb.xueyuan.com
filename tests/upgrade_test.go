package test

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/astaxie/beego/logs"
	gomail "gopkg.in/gomail.v2"
)

//copyfile
func TestCopy(t *testing.T) {
	dstName := "maiyajia.com.exe"
	srcName := "tmp/maiyajia.com.exe"
	copyFile(dstName, srcName)
	fmt.Printf("复制完成")
}
func copyFile(dstName, srcName string) (written int64, err error) {
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

func TestOutmessage(t *testing.T) {
	type message struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		// Progress float64 `json:"progress"`
	}
	var out message
	out.Code = 0
	out.Message = "正在下载安装包"
	// out.Progress = 100
	logs.Info("send_msg ing:", out)
	send_msg, _ := json.Marshal(out)
	logs.Info("send_msg ing:", send_msg)
}

func TestCopyDir(t *testing.T) {
	copyDir("D:\\LJB\\gopath\\src\\maiyajia.com\\asset", "D:\\LJB\\gopath\\src\\maiyajia.com\\backup")
	//copyDir("D:/LJB/gopath/src/maiyajia.com/asset", "D:/LJB/gopath/src/maiyajia.com/backup")
}
func copyDir(src string, dest string) {
	src_original := src
	err := filepath.Walk(src, func(src string, f os.FileInfo, err error) error {
		if f == nil {
			return err
		}
		if f.IsDir() {
			//          fmt.Println(f.Name())
			//          copyDir(f.Name(), dest+"/"+f.Name())
		} else {
			dest_new := strings.Replace(src, src_original, dest, -1)
			//dest_new := strings.Replace(strings.Replace(src, src_original, dest, -1), "\\", "/", -1)
			logs.Info("dest_new:::::", dest_new)
			CopyFile(src, dest_new)
		}
		return nil
	})
	if err != nil {
		fmt.Printf("filepath.Walk() returned %v\n", err)
	}
}

func CopyFile(src, dst string) (w int64, err error) {
	srcFile, err := os.Open(src)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer srcFile.Close()
	dst_slices := strings.Split(dst, "\\")
	dst_slices_len := len(dst_slices)
	dest_dir := ""
	for i := 0; i < dst_slices_len-1; i++ {
		dest_dir = dest_dir + dst_slices[i] + "/"
	}
	b, err := PathExists(dest_dir)
	if b == false {
		err := os.Mkdir(dest_dir, os.ModePerm) //在当前目录下生成md目录
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
func TestRemoveDir(t *testing.T) {
	os.Rename("../tmp/medal", "../asset/medal")
}

// func TestTraction(t *testing.T) {
// 	// 	o := NewOrm()
// 	// err := o.Begin()
// 	// // 事务处理过程

// 	// // 此过程中的所有使用 o Ormer 对象的查询都在事务处理范围内
// 	// if SomeError {
// 	//     err = o.Rollback()
// 	// } else {
// 	//     err = o.Commit()
// }
func TestTime(t *testing.T) {
	// today, _ := time.Parse("2006-01-02", time.Now().Format("2006-01-02"))
	// logs.Info("today:", today)
	// logs.Info("time.Now():", time.Now())
	// theMon := time.Date(2018, time.Month(9), 1, 0, 0, 0, 0, time.Local)
	// monStart := theMon.AddDate(0, 0, 0)
	// monEnd := theMon.AddDate(0, 1, 0)
	// logs.Info("monStart:", monStart)
	// logs.Info("monEnd:", monEnd)
	flag := make(chan int)
	go func() {
		// 	time.Sleep(time.Second * 5)
		flag <- 1
	}()
	logs.Info("flag1:")
	logs.Info("flag2:", <-flag)
}
func TestMail(t *testing.T) {
	gomail.NewMessage()
}
