// @APIVersion 1.0.0
// @Title 系统路由表
// @Description 系统路由信息表
// @Contact xuchuangxin@icanmake.cn
// @TermsOfServiceUrl https://maiyajia.com/
// @License
// @LicenseUrl
package routers

import (
	"maiyajia.com/controllers"

	"github.com/astaxie/beego"
)

func init() {
	ns := beego.NewNamespace("/api",
		beego.NSNamespace("/passport",
			beego.NSRouter("/login", &controllers.PassportController{}, "post:Login"),
			beego.NSRouter("/register", &controllers.PassportController{}, "post:Register"),
			beego.NSRouter("/chanpass", &controllers.PassportController{}, "post:ChangePass"),
			//用户邮箱通行接口
			beego.NSRouter("/email/identification", &controllers.PassportController{}, "put:BindEmail"),
			beego.NSRouter("/email/activation", &controllers.PassportController{}, "get:ActivateEmail"),
			beego.NSRouter("/retrieve", &controllers.PassportController{}, "get:RetrievePassMail"),
			beego.NSRouter("/mailvalid", &controllers.PassportController{}, "get:MailValid"),
			beego.NSRouter("/resetpass", &controllers.PassportController{}, "post:ResetPass"),
		),
		beego.NSNamespace("/admin",
			beego.NSRouter("/repass/student", &controllers.PassportController{}, "post:ResetStudentPass"),
			beego.NSRouter("/repass/teacher", &controllers.PassportController{}, "post:ResetTeacherPass"),
			beego.NSRouter("/blocked", &controllers.PassportController{}, "post:BlockedAccount"),

			//查询系统的信息
			beego.NSRouter("/system/info", &controllers.AdminController{}, "get:GetSystemInfo"),
			beego.NSRouter("/system/account", &controllers.AdminController{}, "get:GetSystemAccount"),
			//系统升级检查
			beego.NSRouter("/system/upgrade/check", &controllers.AdminController{}, "get:CheckSystemUpgrade"),
			beego.NSRouter("/system/upgrade/launch", &controllers.AdminController{}, "get:LaunchSystemUpgrade"),
			beego.NSRouter("/system/upgrade/reboot", &controllers.AdminController{}, "get:RebootSystemUpgrade"),
			//申请的审核接口
			beego.NSRouter("/application/list", &controllers.AdminController{}, "get:GetApplicationlist"),
			beego.NSRouter("/application/review", &controllers.AdminController{}, "put:ReviewApplication"),
			//用户活跃度
			beego.NSRouter("/user/liveness/:startyear:int/:startmonth:int/:endyear:int/:endmonth:int", &controllers.AdminController{}, "get:GetLivenessCount"),
			beego.NSRouter("/user/liveness", &controllers.AdminController{}, "get:InsertLiveness"),

			//线上平台下载课程工具（未使用）
			beego.NSRouter("/data", &controllers.AdminController{}, "get:DownloadData"),
		),
		beego.NSNamespace("/user",
			beego.NSRouter("/checktoken", &controllers.PublicController{}, "get:CheckToken"),
			// 查询用户
			beego.NSRouter("/query/:userid:string", &controllers.PublicController{}, "get:QueryUser"),
			// 返回随机头像
			beego.NSRouter("/avatar/random", &controllers.PublicController{}, "get:RandomAvatar"),
			//用户身份申请提交接口
			beego.NSRouter("/application/submit", &controllers.PublicController{}, "post:SubmitApplication"),
			beego.NSRouter("/application/progress", &controllers.PublicController{}, "get:GetApplicationProgress"),
			// 更新用户头像
			beego.NSRouter("/avatar/update", &controllers.UserController{}, "post:UpdateAvatar"),

			// 用户签到
			beego.NSRouter("/signed", &controllers.SignController{}, "post:Signed"),
			beego.NSRouter("/signedRecords/:year:int/:month:int", &controllers.SignController{}, "get:SignedRecord"),

			// 用户对班级的管理
			beego.NSRouter("/class/join", &controllers.UserController{}, "post:JoinClass"),
			beego.NSRouter("/class/quit", &controllers.UserController{}, "post:QuitClass"),

			// 班级信息公共查询接口
			beego.NSRouter("/class/joined/query/:username:string", &controllers.PublicController{}, "get:QueryJoinClasses"),
			beego.NSRouter("/class/created/query/:id:string", &controllers.PublicController{}, "get:QueryCreatedClasses"),

			// 用户的消息
			beego.NSRouter("/message/query/unread", &controllers.UserController{}, "get:UnreadMessageCount"),
			beego.NSRouter("/message/query/count", &controllers.UserController{}, "get:GetMessageCount"),
			beego.NSRouter("/message/read/new", &controllers.UserController{}, "post:ReadNewMessage"),
			beego.NSRouter("/message/query/:page:int/:number:int", &controllers.UserController{}, "get:QueryMessages"),
			beego.NSRouter("/message/delete", &controllers.UserController{}, "delete:DeleteMessages"),

			// 用户积分排行榜
			beego.NSRouter("/rank/calculus", &controllers.PublicController{}, "get:GetCalculusRank"),

			// 用户学习进度
			beego.NSRouter("/progress", &controllers.ProgressController{}, "post:UploadLessProgress"),
			//获取课时完成状态
			beego.NSRouter("/progress", &controllers.ProgressController{}, "get:GetCourseProgress"),
		),

		beego.NSNamespace("/class",
			beego.NSRouter("/query/:code:string", &controllers.PublicController{}, "get:QueryClass"),
			//beego.NSRouter("/rank/calculus/:code:string", &controllers.PublicController{}, "get:GetClassCalculusRank"),

			// 班级的管理功能，这些接口只对教师或管理员开放
			beego.NSRouter("/create", &controllers.ClassController{}, "post:RegisterClass"),
			beego.NSRouter("/rename", &controllers.ClassController{}, "post:RenameClass"),
			beego.NSRouter("/invite", &controllers.ClassController{}, "post:Invite"),
			beego.NSRouter("/kickout", &controllers.ClassController{}, "post:KickOut"),
			beego.NSRouter("/destroy", &controllers.ClassController{}, "post:DestroyClass"),

			// 班级消息接口，这些接口只对教师或管理员开放
			beego.NSRouter("/message/publish", &controllers.ClassController{}, "post:PublishMessage"),
		),
		beego.NSNamespace("/courses",
			//获取所有课程列表
			beego.NSRouter("/:page:int/:number:int", &controllers.CouresController{}, "get:GetCourses"),
			beego.NSRouter("/category", &controllers.CouresController{}, "get:GetCoursesCategory"),
			beego.NSRouter("/category/:page:int/:number:int", &controllers.CouresController{}, "get:GetCoursesByCategory"),
			//获取指定课程的课时列表
			beego.NSRouter("/:courseId/lessions", &controllers.CouresController{}, "get:GetLessions"),
			//beego.NSRouter("/upgrade/check", &controllers.CouresController{}, "get:CheckCourses"),
			beego.NSRouter("/upgrade/launch", &controllers.CouresController{}, "get:LaunchCourses"),
			//增加老师自定义课程
			beego.NSRouter("/manager/add", &controllers.CouresController{}, "post:AddCustomCourse"),
			beego.NSRouter("/manager/query", &controllers.CouresController{}, "get:GetCustomCourse"),
			beego.NSRouter("/manager/delete", &controllers.CouresController{}, "post:RemoveCustomCourse"),
			beego.NSRouter("/manager/edit", &controllers.CouresController{}, "post:EditCustomCourse"),
			beego.NSRouter("/manager/class", &controllers.CouresController{}, "get:GetClassCourse"),
			//老师获取指定课时下班级学生学习进度
			beego.NSRouter("/lesson/students/progress", &controllers.ProgressController{}, "get:GetStudentsProgress"),
		),
		beego.NSNamespace("/tools",
			//获取所有软件列表
			beego.NSRouter("/", &controllers.ToolsController{}, "get:GetTools"),
			beego.NSRouter("/category", &controllers.ToolsController{}, "get:GetToolsCategory"),
			beego.NSRouter("/weight/:page:int/:number:int", &controllers.ToolsController{}, "get:GetToolsByWeight"),
			beego.NSRouter("/:name", &controllers.ToolsController{}, "get:GetToolByName"),
			//beego.NSRouter("/upgrade/check", &controllers.ToolsController{}, "get:CheckTools"),
			beego.NSRouter("/upgrade/launch", &controllers.ToolsController{}, "get:LaunchTools"),
		),
		beego.NSNamespace("/works",
			//获取个人所有作品列表
			beego.NSRouter("/user/:userid/:page:int/:number:int", &controllers.WorksController{}, "get:GetList"),
			beego.NSRouter("/description", &controllers.WorksController{}, "get:GetDesc"),
			beego.NSRouter("/data", &controllers.WorksController{}, "post:PostBinaryData"),
			beego.NSRouter("/description", &controllers.WorksController{}, "post:PostDescription"),
			beego.NSRouter("/data", &controllers.WorksController{}, "put:PutBinaryData"),
			beego.NSRouter("/description", &controllers.WorksController{}, "put:PutDescription"),
			beego.NSRouter("/", &controllers.WorksController{}, "delete:DeleteWork"),
			beego.NSRouter("/", &controllers.WorksController{}, "get:GetWork"),
			beego.NSRouter("/preview", &controllers.WorksController{}, "get:GetPreview"),
			beego.NSRouter("/share", &controllers.WorksController{}, "put:ShareWork"),
			beego.NSRouter("/sharelist/:page:int/:number:int", &controllers.WorksController{}, "get:GetShareList"),
			beego.NSRouter("/sharelist/category/:page:int/:number:int", &controllers.WorksController{}, "get:GetShareListByCategory"),
			beego.NSRouter("/sharelist/person/:page:int/:number:int", &controllers.WorksController{}, "get:GetPersonalShareList"),
			beego.NSRouter("/laud", &controllers.WorksController{}, "post:GiveLaud"),
			beego.NSRouter("/count", &controllers.WorksController{}, "get:GetLaudAndFavorCou"),
			beego.NSRouter("/favorite", &controllers.WorksController{}, "post:FavoriteWork"),
			beego.NSRouter("/favorite/:page:int/:number:int", &controllers.WorksController{}, "get:GetMyFavoriteWorks"),
			beego.NSRouter("/favorite", &controllers.WorksController{}, "delete:DeleteFavoriteWork"),
			beego.NSRouter("/copywork", &controllers.WorksController{}, "get:CopyWork"),
			beego.NSRouter("/browse", &controllers.WorksController{}, "put:RecordBrowse"),
			beego.NSRouter("/scratch", &controllers.WorksController{}, "post:SaveProject"),
			beego.NSRouter("/project/:page:int/:number:int", &controllers.WorksController{}, "get:LoadProject"),
			beego.NSRouter("/3d-one", &controllers.WorksController{}, "post:Save3DOne"),
			beego.NSRouter("/download", &controllers.WorksController{}, "get:DownloadWork"),
			//**获取指定班级课节下学生作品
			beego.NSRouter("/class/students", &controllers.WorksController{}, "get:GetClassStudents"),
		),
		beego.NSNamespace("/exercise",
			beego.NSRouter("/", &controllers.ExercisesController{}, "post:SaveExer"),
		),
		beego.NSNamespace("/medal",
			//查询全部勋章
			beego.NSRouter("/all", &controllers.PublicController{}, "get:GetAllMedals"),
		),
		/*********** 一下为测试接口， 发布版本是需要删除 ***********/

	)
	beego.AddNamespace(ns)
}
