package course

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/astaxie/beego"
	"maiyajia.com/models"
	"maiyajia.com/services/daemon"
	"maiyajia.com/services/mongo"
)

type CourseRqueryResult struct {
	Code    int                     `json:"code"`
	Courses []models.CourseCategory `json: "courses"`
}

//执行获取所有课程
func execGetAllCourses() bool {
	courseCategories := getAllCourses()
	if courseCategories == nil {
		return false
	}
	if downloadCourses(courseCategories) == false {
		return false
	}
	//记录数据库
	var mgoClient mongo.MgoClient
	mgoClient.StartSession()

	return true
}

// 从课程商店平台获取所有课程
func getAllCourses() (courses []models.CourseCategory) {
	courseStoreUrl := beego.AppConfig.String("course_store_url")
	key, _, _ := daemon.GetProductInfo()
	response, err := http.Post(courseStoreUrl, "application/json", strings.NewReader(key))
	if err == nil {
		return nil
	}
	if response.StatusCode != 200 {
		return nil
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil
	}
	var result CourseRqueryResult
	json.Unmarshal(body, &result)

	if result.Code == 0 {
		return result.Courses
	}
	return nil
}

//下载课程文件
func downloadCourses(courseCategories []models.CourseCategory) bool {
	downloadDir := "./courses"
	for _, catelogy := range courseCategories {
		for _, course := range catelogy.Courses {
			//1. 下载
			res, err := http.Get(course.DownloadURL)
			if err != nil {
				//记录下载失败
				return false
			}
			urlPart := strings.Split(course.DownloadURL, "/")
			downloadFileName := urlPart[len(urlPart)-1]
			targetPath := filepath.Join(downloadDir, downloadFileName)

			f, err := os.Create(targetPath)
			if err != nil {
				//记录失败原因
				return false
			}
			//2.拷贝到指定目录
			io.Copy(f, res.Body)
			// //3.解压
			// unzipPath := filepath.Join(downloadDir, course.ID.Hex())
			// util.DeCompress(targetPath, unzipPath)
			// //记录解压成功
			// course.DownloadSuccess = true
		}
	}
	return true
}
