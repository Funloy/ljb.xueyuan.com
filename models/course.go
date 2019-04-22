package models

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/cavaliercoder/grab"
	"github.com/gorilla/websocket"
	"github.com/mholt/archiver"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"maiyajia.com/services/mongo"
	"maiyajia.com/util"
)

type CourseModels struct {
	MgoSession *mongo.MgoClient
}

const (
	CourseStatusAll       = "all"
	CourseStatusPurchased = "purchased"
	CourseStatusInstalled = "installed"
)

type UpgradeCourses struct {
	Code    int      `json:"code"`
	Newver  bool     `json:"newver"`
	Courses []Course `json:"courses"`
}

//CourseWithLession 指定课程下的部分课程信息及课时列表
type CourseWithLession struct {
	ID       bson.ObjectId `bson:"_id" json:"id"`
	Name     string        `bson:"name" json:"name"` //课程名称
	Icon     string        `bson:"icon" json:"icon"` //icon的URL
	Desc     string        `bson:"desc" json:"desc"` //课程描述
	Lessions []Lession     `bson:"lessions" json:"lessions"`
}

// CoursePreview 课程预览结构
type CoursePreview struct {
	ID       bson.ObjectId `bson:"_id" json:"id"`
	Icon     string        `bson:"icon" json:"icon"`         //icon的URL
	Name     string        `bson:"name" json:"name"`         //课程名称
	Category string        `bson:"category" json:"category"` //所属类别

}

type PatchBody struct {
	CourseID string         `bson:"courseID" json:"courseID"`
	Class    []ClassPreview `bson:"class" json:"class"`
}

// ClassPreview 课程预览信息
type ClassPreview struct {
	Name string `bson:"name" json:"name"`
	Code string `bson""code",json:"code"`
}

// CustomCourse 老师选定课程
type CustomCourse struct {
	ID     bson.ObjectId  `bson:"_id"          json:"id"`
	Class  []ClassPreview `bson:"class"      json:"class"`
	UserID bson.ObjectId  `bson:"userID" json:"userID"`
	Course CoursePreview  `bson:"course" json:"course"`
	Date   time.Time      `bson:"date" json:"date"`
}

// CourseCategory 课程按类别分类
type CourseCategory struct {
	// ID      string   `bson:"_id" json:"id"`
	Name    string   `bson:"name" json:"category_name"` //
	Courses []Course `bson:"courses" json:"courses"`    //课程列表
}

//LessListRes 课程内容列表
type LessListRes struct {
	Code    int      `json:"code"`
	Courses []Course `json:"courses"`
}

// Course 课程
type Course struct {
	ID          bson.ObjectId `bson:"_id" json:"id"`
	Name        string        `bson:"name" json:"name"`                 //课程名称
	DownloadURL string        `bson:"download_url" json:"download_url"` //下载课程包的URL
	Version     string        `bson:"version" json:"version"`
	Icon        string        `bson:"icon" json:"icon"`           //icon的URL
	Category    string        `bson:"category" json:"category"`   //所属类别
	Desc        string        `bson:"desc" json:"desc"`           //课程描述
	Onsell      bool          `bson:"onsell" json:"onsell"`       //是否下载成功
	Purchased   bool          `bson:"purchased" json:"purchased"` //是否已经购买
	Relpath     string        `bson:"relpath" json:"relpath"`     //
	BasePath    string        `bson:"basepath" json:"base_path"`  //
	Lessions    []*Lession    `bson:"lessions" json:"lessions"`   //课时信息
	CreateTime  int64         `bson:"createTime" json:"createTime"`
	Browse      int           `bson:"browse" json:"browse"` //课程浏览量
}

// Lession 课时,即一节课
type Lession struct {
	ID          bson.ObjectId `bson:"_id" json:"id"`
	Name        string        `bson:"name" json:"name"`                 //课节名称
	IconURL     string        `bson:"icon_url" json:"icon_url"`         //图标路径
	QuestionURL string        `bson:"question_url" json:"question_url"` //答题路径
	Contents    []*Content    `bson:"content" json:"content"`           //课节视频
	Tool        string        `bson:"tool" json:"tool"`
}

//Content 学习资源信息
type Content struct {
	ID        bson.ObjectId `bson:"_id" json:"id"`
	VideoName string        `bson:"video_name" json:"video_name"` //视频名称
	VideoURL  string        `bson:"video_url" json:"video_url"`   //视频路径
	MdURL     string        `bson:"md_url" json:"md_url"`         //课时md路径
}

//RecordBrowseOfCourse 记录作品浏览量
func (courseMod *CourseModels) RecordBrowseOfCourse(courseID string) error {
	f := func(col *mgo.Collection) error {
		query := bson.M{"_id": bson.ObjectIdHex(courseID)}
		change := mgo.Change{
			Update:    bson.M{"$inc": bson.M{"browse": 1}},
			ReturnNew: true,
		}
		_, err := col.Find(query).Apply(change, nil)
		return err
	}
	err := courseMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "course", f)
	return err
}

// GetAllCourses 根据是否已安装/是否已购买/所有获取所有的课程列表
func (courseMod *CourseModels) GetAllCourses(paging PagingInfo) (interface{}, error) {
	var courses []Course
	offset := paging.Offset()
	limit := paging.Limit()
	f := func(col *mgo.Collection) error {
		return col.Find(nil).Select(bson.M{"lessions": 0}).Sort("-browse").Limit(limit).Skip(offset).All(&courses)
	}
	// f := func(col *mgo.Collection) error {
	// 	pipeline := []bson.M{
	// 		{"$sort": bson.M{"browse": -1}},
	// 		{"$project": bson.M{"lessions": 0}},
	// 		{"$group": bson.M{"_id": "$category", "name": bson.M{"$first": "$category"}, "courses": bson.M{"$push": "$$ROOT"}}},
	// 	}
	// 	return col.Pipe(pipeline).All(&courses)
	// }

	err := courseMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "course", f)
	return courses, err
}

//GetCoursesInfo 获取所有课程信息，用于升级新增
func (courseMod *CourseModels) GetCoursesInfo() (interface{}, error) {
	var courses []bson.M
	f := func(col *mgo.Collection) error {
		pipeline := []bson.M{
			{"$project": bson.M{"_id": 1, "name": 1, "version": 1}},
		}
		return col.Pipe(pipeline).All(&courses)
	}
	err := courseMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "course", f)
	logs.Debug("Begin getTools in models:", err)
	return courses, err
}

// InstallCourses 接口在线获取课程并下载，写入课程信息到本地
func (courseMod *CourseModels) InstallCourses(url string, productKey string, productSerial string, ws *websocket.Conn) error {
	logs.Info("begin Install course")
	// 获取所有的课程信息
	mongo.Client = &mongo.MgoClient{}
	mongo.Client.StartSession()
	defer mongo.Client.CloseSession()
	CourseList, err := receiveAllCourses(url, productKey, productSerial)
	if err != nil {
		logs.Info("receiveAllCourses fail:", err)
		return err
	}
	//清空数据库课程信息
	if err := courseMod.clearCourseData(); err != nil {
		logs.Info("clearCourseData error:", err)
		return err
	}
	//多文件下载
	courseMod.DownloadCourses(CourseList, ws)
	logs.Info("end Install course")
	return nil
}

//DownloadCourses 接口在线多文件下载
func (courseMod *CourseModels) DownloadCourses(courses []Course, ws *websocket.Conn) error {
	var out Out
	logs.Info("Begin Download Courses")
	//判断是否购买
	reqs := make([]*grab.Request, 0)
	//判断是否为七牛云公开下载
	isQINIU := beego.AppConfig.DefaultBool("isQINIU", false)

	if isQINIU {
		for _, cou := range courses {
			if cou.Purchased == false {
				continue
			}
			logs.Info("课程名字：", cou.Name)
			if err := courseMod.InsertCourse(cou); err != nil {
				logs.Info("InsertCourse(course)", err)
			}
		}
		////提示下载完成
		out.Code = 0
		out.State = &WSProgress{
			Stage:    DOWNLOAD,
			Progress: 100,
		}
		send_msg, err := json.Marshal(out)
		ws.WriteMessage(websocket.TextMessage, []byte(send_msg))
		if err != nil {
			logs.Error("error:", err)
		}
	} else {
		token, err := util.NetdiskAuth()
		if err != nil {
			logs.Error("网盘登陆失败：", err)
			return err
		}
		for _, cou := range courses {
			if cou.Purchased == false {
				continue
			}
			logs.Info("课程名字：", cou.Name)
			req, _ := grab.NewRequest(path.Join(beego.AppPath, "asset", "courses", cou.ID.Hex(), "course.tar"), cou.DownloadURL)
			req.HTTPRequest.Header.Add("Authorization", "Bearer "+string(token))
			req.Tag = cou
			reqs = append(reqs, req)
		}
		// start files downloads, arg0 at a time
		respch := grab.DefaultClient.DoBatch(0, reqs...)
		t := time.NewTicker(200 * time.Millisecond)
		// monitor downloads
		completed := 0
		inProgress := 0
		responses := make([]*grab.Response, 0)
		for completed < len(reqs) {
			select {
			case resp := <-respch:
				// a new response has been received and has started downloading
				// (nil is received once, when the channel is closed by grab)
				if resp != nil {
					responses = append(responses, resp)
				}

			case <-t.C:
				// update completed downloads
				inProgress = 0
				for i, resp := range responses {
					// update downloads in progress
					if resp != nil {
						inProgress++
						out.Code = 0
						out.State = &WSProgress{
							Stage:    DOWNLOAD,
							Name:     resp.Request.Tag.(Course).Name,
							Progress: int(100 * resp.Progress()),
						}
						send_msg, err := json.Marshal(out)
						if err != nil {
							logs.Error("error:", err)
						}
						ws.WriteMessage(websocket.TextMessage, []byte(send_msg))
						logs.Info("Transferred %v/%v bytes (%.2f)", resp.BytesComplete(), resp.Size, 100*resp.Progress())
					}
					if resp != nil && resp.IsComplete() {
						////提示下载完成
						out.Code = 0
						out.State = &WSProgress{
							Stage:    DOWNLOAD,
							Name:     resp.Request.Tag.(Course).Name,
							Progress: 100,
						}
						send_msg, err := json.Marshal(out)
						ws.WriteMessage(websocket.TextMessage, []byte(send_msg))
						if err != nil {
							logs.Error("error:", err)
						}
						// print final result
						course := resp.Request.Tag.(Course)
						if resp.Err() != nil {
							logs.Info("Download failed: %v\n", resp.Err())
							return resp.Err()
						} else {

							logs.Info("Download saved to ./%v \n", resp.Filename)
							// 解压安装包
							out.Code = 0
							out.State = &WSProgress{
								Stage:    UNCOMPRESSION,
								Name:     resp.Request.Tag.(Course).Name,
								Progress: 0,
							}
							send_msg, err := json.Marshal(out)
							if err != nil {
								logs.Error("error:", err)
							}
							ws.WriteMessage(websocket.TextMessage, []byte(send_msg))
							//Open解压文件到指定文件夹asset
							if error := archiver.Tar.Open(path.Join(beego.AppPath, "asset", "courses", course.ID.Hex(), "course.tar"), path.Join(beego.AppPath, "asset", "courses", course.ID.Hex())); error != nil {
								logs.Info("File extract fail!", error)
								return error
							}
							if err := courseMod.handleCourse(course); err != nil {
								return err
							}
							// 解压安装包
							out.Code = 0
							out.State = &WSProgress{
								Stage:    UNCOMPRESSION,
								Name:     resp.Request.Tag.(Course).Name,
								Progress: 100,
							}
							send_msg, err = json.Marshal(out)
							ws.WriteMessage(websocket.TextMessage, []byte(send_msg))
						}
						// mark completed
						responses[i] = nil
						completed++
					}
				}
			}

		}
		out.Code = 0
		out.State = &WSProgress{
			Stage: DOEN,
		}
		send_msg, _ := json.Marshal(out)
		ws.WriteMessage(websocket.TextMessage, []byte(send_msg))
		t.Stop()
		logs.Info("%d files successfully downloaded.\n", completed)
	}
	return nil
}

//InsertCourse 写入下载课程信息
func (courseMod *CourseModels) InsertCourse(cou Course) error {
	f := func(col *mgo.Collection) error {
		query := bson.M{"name": cou.Name}
		change := mgo.Change{
			Update:    cou,
			ReturnNew: true,
			Upsert:    true,
		}
		_, err := col.Find(query).Apply(change, nil)
		return err
	}
	return courseMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "course", f)

}

//QueryLessonsList 返回已下载课程清单
func (courseMod *CourseModels) QueryLessonsList() ([]Course, error) {
	var lesson []Course
	f := func(col *mgo.Collection) error {
		return col.Find(nil).All(&lesson)
	}
	err := courseMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "course", f)
	return lesson, err
}

// GetAllLessions 获取指定课程下的课时列表
func (courseMod *CourseModels) GetAllLessions(courseId bson.ObjectId) (CourseWithLession, error) {
	logs.Info("courseId,course:", courseId)
	var course CourseWithLession
	f := func(col *mgo.Collection) error {
		return col.Find(bson.M{"_id": courseId}).One(&course)
	}

	err := courseMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "course", f)

	return course, err
}

//GetLessonById 获取指定课时信息
func (courseMod *CourseModels) GetLessonById(courseID, lessonID string) (*Lession, error) {
	var course Course
	f := func(col *mgo.Collection) error {
		return col.Find(bson.M{"_id": bson.ObjectIdHex(courseID)}).One(&course)
	}
	err := courseMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "course", f)
	for _, lesson := range course.Lessions {
		if lesson.ID.Hex() == lessonID {
			logs.Info("lesson:", lesson)
			return lesson, err
		}
	}
	return nil, err
}

// OrderInstallCourses 后台指令获取课程并下载，写入课程信息到本地
func (courseMod *CourseModels) OrderInstallCourses(url string, productKey string, productSerial string) error {
	logs.Info("begin Install course")
	// 获取所有的课程信息
	CourseList, err := receiveAllCourses(url, productKey, productSerial)
	if err != nil {
		logs.Info("receiveAllCourses fail:", err)
		return err
	}
	// logs.Info("CourseList:", CourseList)
	//清空数据库课程信息
	if err := courseMod.clearCourseData(); err != nil {
		logs.Info("clearCourseData error:", err)
		return err
	}
	//判断是否为七牛云公开下载
	isQINIU := beego.AppConfig.DefaultBool("isQINIU", false)
	if isQINIU {
		//线上版直接写入数据库
		for _, course := range CourseList {
			if course.Purchased == false {
				continue
			}
			if err := courseMod.InsertCourse(course); err != nil {
				logs.Info("InsertCourse(course)", err)
			}
		}
	} else {
		// 清空本地课程文件夹
		if err := os.RemoveAll(path.Join(beego.AppPath, "asset", "courses")); err != nil {
			logs.Error("remove dir[%s] fail", path.Join(beego.AppPath, "asset", "courses"))
			return err
		}
		//局域网下载
		courseMod.OrderDownloadCourses(CourseList)
	}
	logs.Info("end Install course")
	return nil
}

//OrderDownloadCourses 局域网后台命令行多文件下载
func (courseMod *CourseModels) OrderDownloadCourses(courses []Course) error {
	logs.Info("Begin Download Courses")
	//判断是否购买
	reqs := make([]*grab.Request, 0)
	token, err := util.NetdiskAuth()
	if err != nil {
		logs.Error("网盘登陆失败：", err)
		return err
	}
	for _, cou := range courses {
		if cou.Purchased == false {
			continue
		}
		logs.Info("课程名字：", cou.Name)
		name := cou.Name + ".zip"
		//req, _ := grab.NewRequest(path.Join(beego.AppPath, "asset", "courses", cou.ID.Hex(), "course.tar"), cou.DownloadURL)
		req, _ := grab.NewRequest(path.Join(beego.AppPath, "asset", "courses", name), cou.DownloadURL)
		req.HTTPRequest.Header.Add("Authorization", "Bearer "+string(token))
		req.Tag = cou
		reqs = append(reqs, req)
	}
	// start files downloads, arg0 at a time
	respch := grab.DefaultClient.DoBatch(0, reqs...)
	t := time.NewTicker(200 * time.Millisecond)
	// monitor downloads
	completed := 0
	inProgress := 0
	responses := make([]*grab.Response, 0)
	for completed < len(reqs) {
		select {
		case resp := <-respch:
			// a new response has been received and has started downloading
			// (nil is received once, when the channel is closed by grab)
			if resp != nil {
				responses = append(responses, resp)
			}

		case <-t.C:
			// update completed downloads
			inProgress = 0
			for i, resp := range responses {
				// update downloads in progress
				if resp != nil {
					inProgress++
					logs.Info("Transferred %v/%v bytes (%.2f)", resp.BytesComplete(), resp.Size, 100*resp.Progress())
				}
				if resp != nil && resp.IsComplete() {
					// print final result
					course := resp.Request.Tag.(Course)
					if resp.Err() != nil {
						logs.Info("Download failed: %v\n", resp.Err())
						return resp.Err()
					} else {
						logs.Info("Download saved to ./%v \n", resp.Filename)
						//Open解压文件到指定文件夹asset
						if error := archiver.Zip.Open(resp.Filename, path.Join(beego.AppPath, "asset", "courses")); error != nil {
							logs.Info("File extract fail!", error)
							return error
						}
						if err := courseMod.handleCourse(course); err != nil {
							return err
						}
					}
					// mark completed
					responses[i] = nil
					completed++
				}
			}
		}

	}
	t.Stop()
	logs.Info("%d files successfully downloaded.\n", completed)
	return nil
}

/*********************************************************************************************/
/*********************************** 以下为本控制器的内部函数 *********************************/
/*********************************** *********************************************************/
//获取所有的课程信息
func receiveAllCourses(url, productKey, productSerial string) ([]Course, error) {
	var courses LessListRes
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logs.Error("NewRequest erro", err)
		return nil, err
	}
	req.Header.Set("productKey", productKey)
	req.Header.Set("productSerial", productSerial)
	resp, err := client.Do(req)
	if err != nil {
		logs.Error("resp Error", err)
		return nil, err
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	courseJSON := buf.String()
	//logs.Info("courseJSON:", courseJSON)
	err = json.Unmarshal([]byte(courseJSON), &courses)
	if err != nil {
		logs.Info("unmarshal json err:", err)
		return nil, err
	}
	return courses.Courses, nil
}

func (courseMod *CourseModels) getUserCouById(courseId string) (Course, error) {
	var course Course
	f := func(col *mgo.Collection) error {
		return col.Find(bson.M{"_id": bson.ObjectIdHex(courseId)}).One(&course)
	}
	err := courseMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "course", f)
	return course, err
}

//handleCourse 写入数据库前对course进行处理
func (courseMod *CourseModels) handleCourse(course Course) error {
	logs.Info("File extract done")
	var (
	//relpath string
	// lesPath string
	// name    string
	// //	mdUrl   string
	// iconUrl string
	// sortArr []int //排序数组
	// lesson  *Lession
	)
	//nameMap := make(map[int]string)
	// indexPath := path.Join(beego.AppPath, "asset", "courses", course.ID.Hex())
	// files, _ := ioutil.ReadDir(indexPath)
	// for _, f := range files {
	// 	if f.IsDir() {
	// 		relpath = path.Join("asset", "courses", course.ID.Hex(), f.Name())
	// 		dirList, _ := ioutil.ReadDir(path.Join(indexPath, f.Name()))
	// 		for _, f := range dirList {
	// 			//3.文件名排序
	// 			if f.IsDir() {
	// 				//截取后三位转换为int型
	// 				cut3, _ := strconv.Atoi(f.Name()[len(f.Name())-3:])
	// 				//k:cut3,v:文件名
	// 				nameMap[cut3] = f.Name()
	// 				sortArr = append(sortArr, cut3)

	// 			} else if strings.HasSuffix(f.Name(), ".png") || strings.HasSuffix(f.Name(), ".jpg") {
	// 				course.Icon = relpath + "/" + f.Name()
	// 			}

	// 		}
	// 		sort.Ints(sortArr)
	// 		for _, num := range sortArr {

	// 			name = nameMap[num]
	// 			//课程目录级
	// 			lesPath = path.Join(relpath, name)
	// 			files, _ := ioutil.ReadDir(lesPath)
	// 			for _, ff := range files {
	// 				logs.Info("ff.Name():", ff.Name())
	// 				if strings.HasSuffix(ff.Name(), ".png") || strings.HasSuffix(ff.Name(), ".jpg") {
	// 					iconUrl = path.Join(lesPath, ff.Name())
	// 				} else if strings.HasSuffix(ff.Name(), ".md") {
	// 					//mdUrl = path.Join(lesPath, ff.Name())
	// 				}
	// 			}
	// 			lesson = &Lession{
	// 				ID:      bson.NewObjectId(),
	// 				Name:    name,
	// 				IconURL: iconUrl,
	// 			}
	// 			course.Lessions = append(course.Lessions, lesson)
	// 		}
	// 		logs.Info("course.Lessions", course.Lessions)
	// 	}
	// }
	//course.Relpath = relpath
	//4. 课程信息写入数据库
	logs.Info("course:", course.Lessions)
	recourse := ReviseCourse(course)
	if err := courseMod.InsertCourse(recourse); err != nil {
		logs.Info("InsertCourse(course)", err)
	}
	name := course.Name + ".zip"
	if error := os.Remove(path.Join(beego.AppPath, "asset", "courses", name)); error != nil {
		//如果删除失败则输出 file remove Error!
		logs.Info("zip file remove Error!")
		return error
	} else {
		//如果删除成功则输出 file remove OK!
		logs.Info("zip file remove OK!")
	}
	return nil
}
func (courseMod *CourseModels) clearCourseData() error {
	f := func(col *mgo.Collection) error {
		_, err := col.RemoveAll(bson.M{})
		return err
	}
	return courseMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "course", f)
}

// CustomCourseParam 自定义课程选取参数结构
type CustomCourseParam struct {
	UserID   string   `bson:"userID" json:"userID"`
	CourseID []string `bson:"courseID" json:"courseID"`
}

// FindCourseByID 根据课程ID查找课程具体信息
func (courseMod *CourseModels) FindCourseByID(id string) (CoursePreview, error) {

	var course CoursePreview

	f := func(col *mgo.Collection) error {

		pipeline := []bson.M{
			{"$project": bson.M{"icon": 1, "name": 1, "category": 1, "_id": 1}},
			{"$match": bson.M{"_id": bson.ObjectIdHex(id)}},
		}
		return col.Pipe(pipeline).One(&course)
		//return col.Find(reqParam).Select(bson.M{"lessions": 0}).All(&courses)
	}
	err := courseMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "course", f)
	return course, err
}

// NewCustomCourses 初始化一个自定义课程数组
func (courseMod *CourseModels) NewCustomCourses(userID string, courseID []string) []*CustomCourse {

	var customCourses []*CustomCourse

	for _, item := range courseID {

		customCourse, err := courseMod.FindCourseByID(item)

		if err == nil {

			var customcourse = &CustomCourse{
				ID:     bson.NewObjectId(),
				Class:  []ClassPreview{},
				UserID: bson.ObjectIdHex(userID),
				Course: CoursePreview{
					ID:       customCourse.ID,
					Icon:     customCourse.Icon,
					Name:     customCourse.Name,
					Category: customCourse.Category,
				},
				Date: time.Now(),
			}

			customCourses = append(customCourses, customcourse)
		}

	}

	return customCourses
}

// RegisteredCustomCourses 在数据库创建一个新的自定义课程
func (courseMod *CourseModels) RegisteredCustomCourses(customCourse interface{}) error {
	f := func(col *mgo.Collection) error {
		return col.Insert(customCourse)
	}
	return courseMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "customcourse", f)
}

// DeleteCustomCourses 删除自定义课程
func (courseMod *CourseModels) DeleteCustomCourses(userID string, courseID string) error {

	f := func(col *mgo.Collection) error {
		return col.Remove(bson.M{"userID": bson.ObjectIdHex(userID), "course._id": bson.ObjectIdHex(courseID)})
	}
	return courseMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "customcourse", f)

}

// PatchCustomCourseClass 修改课程包含班级
func (courseMod *CourseModels) PatchCustomCourseClass(userID string, courseID string, class []ClassPreview) error {

	fmt.Println("patch!!!:", userID, courseID, class)
	f := func(col *mgo.Collection) error {
		return col.Update(bson.M{"userID": bson.ObjectIdHex(userID), "course._id": bson.ObjectIdHex(courseID)}, bson.M{"$set": bson.M{"class": class}})
	}
	return courseMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "customcourse", f)

}

// GetCustomCoursesByUserID 根据用户ID获取课程列表
func (courseMod *CourseModels) GetCustomCoursesByUserID(userID string) ([]CustomCourse, error) {

	var courses []CustomCourse
	f := func(col *mgo.Collection) error {

		pipeline := []bson.M{
			{"$match": bson.M{"userID": bson.ObjectIdHex(userID)}},
		}
		return col.Pipe(pipeline).All(&courses)
	}
	err := courseMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "customcourse", f)
	return courses, err

}

// GetAllCustomCourses 获取所有自定义课程
func (courseMod *CourseModels) GetAllCustomCourses() ([]CustomCourse, error) {

	var course []CustomCourse

	f := func(col *mgo.Collection) error {
		return col.Find(nil).All(&course)
	}

	err := courseMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "customcourse", f)

	return course, err
}

// CustomCourseExit 判断课程是否存在
func (courseMod *CourseModels) CustomCourseExit(userID string, courseID string) (CustomCourse, error) {

	var course CustomCourse
	f := func(col *mgo.Collection) error {
		return col.Find(bson.M{"userID": bson.ObjectIdHex(userID), "course._id": bson.ObjectIdHex(courseID)}).One(&course)
	}

	err := courseMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "customcourse", f)

	return course, err
}
func ReviseCourse(course Course) Course {
	prefix := "asset/course/"
	for _, lesson := range course.Lessions {
		lesson.IconURL = prefix + lesson.IconURL
		lesson.QuestionURL = prefix + lesson.QuestionURL
		for _, content := range lesson.Contents {
			content.VideoURL = prefix + content.VideoURL
			content.MdURL = prefix + content.MdURL
		}
	}
	return course
}

//GetClassCourse 获取班级课程
func (courseMod *CourseModels) GetClassCourse(code string) (interface{}, error) {
	var query []interface{}
	f := func(col *mgo.Collection) error {
		pipeline := []bson.M{
			{"$unwind": "$class"},
			{"$match": bson.M{"class.code": code}},
		}
		return col.Pipe(pipeline).All(&query)
	}
	err := courseMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "customcourse", f)
	return query, err
}

//GetCoursesCategory 获取工具所属类型
func (courseMod *CourseModels) GetCoursesCategory() ([]Category, error) {
	var categorys []Category
	f := func(col *mgo.Collection) error {
		pipeline := []bson.M{
			{"$group": bson.M{"_id": "$category", "category": bson.M{"$first": "$category"}}},
		}
		return col.Pipe(pipeline).All(&categorys)
	}
	err := courseMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "course", f)
	if err != nil {
		logs.Error("Begin GetCoursesCategory in models:", err)
	}
	return categorys, err
}

//QueryCoursesByCategoryCount 查询指定类别下课程总数
func (courseMod *CourseModels) QueryCoursesByCategoryCount(category string) (TotalBody, error) {
	pipeline := []bson.M{
		{"$match": bson.M{"category": category}},
		{"$group": bson.M{"_id": nil, "total": bson.M{"$sum": 1}}},
		{"$project": bson.M{"_id": 0}},
	}
	var total TotalBody
	f := func(col *mgo.Collection) error {
		return col.Pipe(pipeline).One(&total)
	}
	return total, courseMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "course", f)
}

//QueryCoursesCount 查询课程总数
func (courseMod *CourseModels) QueryCoursesCount() (int, error) {
	var total int
	var err error
	f := func(col *mgo.Collection) error {
		total, err = col.Find(nil).Count()
		return err
	}
	return total, courseMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "course", f)
}

//GetCoursesByCategory 查询指定类别下课程信息

func (courseMod *CourseModels) GetCoursesByCategory(paging PagingInfo, category string) (interface{}, error) {
	var courses []Course
	offset := paging.Offset()
	limit := paging.Limit()
	f := func(col *mgo.Collection) error {
		return col.Find(bson.M{"category": category}).Select(bson.M{"lessions": 0}).Sort("-browse").Limit(limit).Skip(offset).All(&courses)
	}
	err := courseMod.MgoSession.Do(beego.AppConfig.String("MongoDB"), "course", f)
	return courses, err
}
