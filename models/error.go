// @APIVersion 1.0.0
// @Title 业务逻辑错误模型
// @Description 定义业务逻辑错误时发生的错误代码和错误描述
// @Contact xuchuangxin@icanmake.cn
// @TermsOfServiceUrl https://maiyajia.com/
// @License
// @LicenseUrl

package models

const (

	// 0 代表成功，非0代表错误
	_ = iota

	ERR_REQUEST_PARAM
	ERR_PERMISSION_DENIED

	ERR_LOGIN_FAIL
	ERR_LOGIN_PWD_FAIL
	ERR_TOKEN_FMT_FAIL
	ERR_LOGIN_TOKEN_FAIL

	ERR_ACTIVATE_TOKEN_FAIL
	ERR_RETRIEVE_TOKEN_FAIL
	ERR_PARSE_ACTIVATETOKEN_FAIL
	ERR_PARSE_REPASSTOKEN_FAIL
	ERR_BIND_EMAIL_FAIL
	ERR_SEND_EMAIL_FAIL
	ERR_EMAIL_NONE
	ERR_EMAIL_EXIST

	ERR_REGISTER_FAIL
	ERR_REGISTER_KEY_FAIL
	ERR_USER_FMT_FAIL
	ERR_USER_NONE
	ERR_USER_EXISTS
	ERR_USER_LOCKED
	ERR_PWD_FMT_FAIL
	ERR_PWD_CHAN_FAIL
	ERR_PWD_CHAN_OLD_FAIL
	ERR_PWD_CHAN_UPDATE_FAIL
	ERR_AVATAR_CHAN_FAIL

	ERR_USER_APPLICATION_SUBMIT_FAIL
	ERR_USER_APPLICATION_SUBMIT_EXSIT
	ERR_USER_APPLICATION_PROGRESS_FAIL
	ERR_USER_APPLICATION_SUCCESS
	ERR_REVIEW_APPLICATION_FAIL
	ERR_DELLIVENESS_FAIL
	ERR_INSERTUSERLIVE_FAIL
	ERR_GETUSERLIVE_FAIL
	ERR_GETUSERSCOUNT_FAIL

	ERR_ADMIN_PERMISSION_ROLE
	ERR_ADMIN_REPASS_FAIL
	ERR_ADMIN_BLOCKED_FAIL
	ERR_ADMIN_UNBLOCKED_FAIL
	ERR_ADMIN_BLOCKED_EXISTS
	ERR_ADMIN_UNBLOCKED_EXISTS
	ERR_ADMIN_UNPERMIT
	ERR_ADMIN_APPLICATION_QUERY_FAIL
	ERR_ADMIN_LIVENESS_QUERY_FAIL

	ERR_SIGN_FAIL
	ERR_SIGN_RECORD_FAIL
	ERR_SIGN_DUP_FAIL

	//积分排行版查询错误
	ERR_CALCULUS_RANK_QUERY_FAIL

	ERR_CLASS_CREATE_FAIL
	ERR_CLASS_QUERY_FAIL
	ERR_CLASS_FMT_FAIL
	ERR_CLASS_NONE
	ERR_CLASS_EXISTS
	ERR_CLASS_CODE_FAIL
	ERR_CLASS_CODE_NONE
	ERR_CLASS_USER_EXISTS
	ERR_CLASS_USER_NONE
	ERR_CLASS_JOIN_FAIL
	ERR_CLASS_QUIT_FAIL
	ERR_CLASS_RENAME_FAIL
	ERR_CLASS_KICK_FAIL
	ERR_CLASS_DESTROY_FAIL
	ERR_CLASS_DESTROY_NOBODY

	ERR_CLASS_MESSAGE_PUBLISH
	ERR_CLASS_MESSAGE_PUBLISH_FAIL
	ERR_CLASS_MESSAGE_READ_FAIL
	ERR_CLASS_MESSAGE_QUERY_FAIL
	ERR_CLASS_MESSAGE_DELETE_FAIL

	//课程查询错误
	ERR_COURSE_MESSAGE_QUERY_FAIL
	ERR_COURSE_DOWNLOAD_FAIL
	ERR_COURSE_BROWSE_FAIL
	//工具查询错误
	ERR_TOOL_MESSAGE_QUERY_FAIL
	ERR_TOOL_DOWNLOAD_FAIL
	//课时查询错误
	ERR_LESSIONS_MESSAGE_QUERY_FAIL

	//更新课节状态错误
	ERR_LESSION_PROGRESS_UPDATE_FAIL
	ERR_LESSION_STATUS_UPDATE_FAIL
	ERR_LESSION_PROGRESS_QUERY_FAIL
	//查询课程学习进展错误
	ERR_COURSE_PROGRESS_QUERY_FAIL

	//平台升级错误
	ERR_UPGRADE_LOGIN_TOKEN
	ERR_UPGRADE_CHECK_FAIL
	ERR_UPGRADE_FAIL
	ERR_UPGRADE_DOWNLOAD
	ERR_UPGRADE_REBOOT
	ERR_UPGRADE_DOEN

	// 作品
	ERR_NO_WORK_EXISTS
	ERR_NO_file_EXISTS
	ERR_SGL_NO_file_EXISTS
	ERR_WORK_EXISTS
	ERR_ADD_WORK_FAIL
	ERR_SHARE_WORK_FAIL
	ERR_DELETE_WORK_FAIL
	ERR_READ_WORK_FAIL
	ERR_WRITE_WORK_FAIL
	ERR_CREATE_FILE_FAIL
	ERR_LAUD_WORK_FAIL
	ERR_LAUD_RECORD_QUERY_FAIL
	ERR_SELF_UNCOLLECTION_FAIL
	ERR_QUERY_FAVOR_FAIL
	ERR_COLLECTION_WORK_FAIL
	ERR_FAVORITE_WORK_LIST
	ERR_DELETE_FAVORITE_WORK
	ERR_UPLOAD_IMAGES_FAIL
	ERR_COUNT_FAIL

	// 练习
	ERR_ADD_EXER_FAIL
	// 选课
	ERR_CUSTOMECOURSE_EXIT
	ERR_CUSTOMECOURSE_NULL
	ERR_CUSTOMECOURSE_DELETE

	ERR_DIRECTORY_CREATE

	ERR_TOOL_REALPATH

	RUNTIME_ERROR
)

// ErrorResult 服务端错误响应
type ErrorResult struct {
	Code    int    `json:"code"`
	Message string `json:"message,omitempty"`
}

var errorMsgs map[int]string

// ErrorMsgs 获取错误表
func ErrorMsgs() map[int]string {
	if errorMsgs == nil {
		errorMsgs = make(map[int]string)

		errorMsgs[ERR_PERMISSION_DENIED] = "权限不足，无法访问"

		errorMsgs[ERR_REQUEST_PARAM] = "请求参数错误"
		errorMsgs[ERR_LOGIN_FAIL] = "登录请求错误"
		errorMsgs[ERR_LOGIN_PWD_FAIL] = "账号或密码错误"
		errorMsgs[ERR_LOGIN_TOKEN_FAIL] = "无法获取登录令牌"

		errorMsgs[ERR_TOKEN_FMT_FAIL] = "token格式错误或token已过期"

		errorMsgs[ERR_REGISTER_FAIL] = "注册请求错误"
		errorMsgs[ERR_REGISTER_KEY_FAIL] = "输入的产品KEY错误"
		errorMsgs[ERR_USER_FMT_FAIL] = "用户名长度4~20位，仅能使用字母，数字，下划线，减号"
		errorMsgs[ERR_PWD_FMT_FAIL] = "密码的长度需在6-16位之间"
		errorMsgs[ERR_PWD_CHAN_FAIL] = "修改密码请求错误"
		errorMsgs[ERR_PWD_CHAN_OLD_FAIL] = "原密码错误"
		errorMsgs[ERR_PWD_CHAN_UPDATE_FAIL] = "更新密码时发生错误"
		errorMsgs[ERR_ACTIVATE_TOKEN_FAIL] = "无法获取邮箱激活令牌"
		errorMsgs[ERR_RETRIEVE_TOKEN_FAIL] = "无法获取密码找回令牌"
		errorMsgs[ERR_SEND_EMAIL_FAIL] = "发送邮件失败，请重新绑定激活"
		errorMsgs[ERR_PARSE_ACTIVATETOKEN_FAIL] = "邮箱激活失败,邮件已过期"
		errorMsgs[ERR_PARSE_REPASSTOKEN_FAIL] = "密码重置链接失效,邮件已过期"
		errorMsgs[ERR_BIND_EMAIL_FAIL] = "邮箱绑定失败"
		errorMsgs[ERR_EMAIL_NONE] = "该邮箱未绑定账号"
		errorMsgs[ERR_EMAIL_EXIST] = "该邮箱已绑定账号"

		errorMsgs[ERR_AVATAR_CHAN_FAIL] = "更新头像失败，请稍后重试"

		errorMsgs[ERR_USER_APPLICATION_SUBMIT_FAIL] = "提交申请失败，请稍后重试"
		errorMsgs[ERR_USER_APPLICATION_SUBMIT_EXSIT] = "已提交申请，请勿重复提交"
		errorMsgs[ERR_USER_APPLICATION_PROGRESS_FAIL] = "获取申请进度失败"
		errorMsgs[ERR_USER_APPLICATION_SUCCESS] = "审核已通过，请勿重复提交"
		errorMsgs[ERR_REVIEW_APPLICATION_FAIL] = "审核提交失败，请稍后重试"
		errorMsgs[ERR_DELLIVENESS_FAIL] = "清空月活越失败，请稍后重试"
		errorMsgs[ERR_INSERTUSERLIVE_FAIL] = "添加用户活跃记录失败，请稍后重试"
		errorMsgs[ERR_GETUSERLIVE_FAIL] = "获取用户月活跃度，请稍后重试"
		errorMsgs[ERR_GETUSERSCOUNT_FAIL] = "获取注册总数，请稍后重试"

		errorMsgs[ERR_USER_LOCKED] = "账号已被锁定，请联系管理员解锁"
		errorMsgs[ERR_USER_NONE] = "用户不存在"
		errorMsgs[ERR_USER_EXISTS] = "用户已被占用"

		errorMsgs[ERR_ADMIN_PERMISSION_ROLE] = "只能管理学生账户"
		errorMsgs[ERR_ADMIN_REPASS_FAIL] = "重置账户的密码失败，请稍后重试"
		errorMsgs[ERR_ADMIN_BLOCKED_EXISTS] = "账号没有被冻结"
		errorMsgs[ERR_ADMIN_UNBLOCKED_EXISTS] = "账号没有被解封"
		errorMsgs[ERR_ADMIN_BLOCKED_FAIL] = "冻结账号失败，请稍后重试"
		errorMsgs[ERR_ADMIN_UNBLOCKED_FAIL] = "解封账号失败，请稍后重试"
		errorMsgs[ERR_ADMIN_UNPERMIT] = "权限不足"
		errorMsgs[ERR_ADMIN_APPLICATION_QUERY_FAIL] = "资格申请列表查询失败，请稍后重试"
		errorMsgs[ERR_ADMIN_LIVENESS_QUERY_FAIL] = "用户月活越查询失败"

		errorMsgs[ERR_SIGN_FAIL] = "签到失败，请稍后重试"
		errorMsgs[ERR_SIGN_DUP_FAIL] = "今天已经签到"
		errorMsgs[ERR_SIGN_RECORD_FAIL] = "查询签到日期失败"

		errorMsgs[ERR_CALCULUS_RANK_QUERY_FAIL] = "暂时无法获取积分排行榜，请稍后重试"

		errorMsgs[ERR_CLASS_CREATE_FAIL] = "创建班级失败，请稍后重试"
		errorMsgs[ERR_CLASS_NONE] = "班级不存在"
		errorMsgs[ERR_CLASS_FMT_FAIL] = "班级名称格式必须含中文"
		errorMsgs[ERR_CLASS_EXISTS] = "班级名称已被占用"
		errorMsgs[ERR_CLASS_CODE_FAIL] = "班级代码错误"
		errorMsgs[ERR_CLASS_CODE_NONE] = "班级代码不存在"
		errorMsgs[ERR_CLASS_USER_EXISTS] = "你已经是班级成员"
		errorMsgs[ERR_CLASS_USER_NONE] = "你不是该班级的成员"
		errorMsgs[ERR_CLASS_JOIN_FAIL] = "加入班级失败，请稍后重试"
		errorMsgs[ERR_CLASS_QUIT_FAIL] = "退出班级失败，请稍后重试"
		errorMsgs[ERR_CLASS_RENAME_FAIL] = "班级创建者才能修改班级名称"
		errorMsgs[ERR_CLASS_KICK_FAIL] = "用户已经不在班级"
		errorMsgs[ERR_CLASS_DESTROY_NOBODY] = "班级学生人数不为零，不能删除班级"
		errorMsgs[ERR_CLASS_DESTROY_FAIL] = "删除班级失败，请稍后重试"
		errorMsgs[ERR_CLASS_QUERY_FAIL] = "查询班级信息错误，请稍后重试"

		errorMsgs[ERR_CLASS_MESSAGE_PUBLISH] = "班级未创建学生"
		errorMsgs[ERR_CLASS_MESSAGE_PUBLISH_FAIL] = "发布班级消息失败，请稍后重试"
		errorMsgs[ERR_CLASS_MESSAGE_READ_FAIL] = "阅读班级消息失败，请稍后重试"
		errorMsgs[ERR_CLASS_MESSAGE_QUERY_FAIL] = "获取班级消息失败，请稍后重试"
		errorMsgs[ERR_CLASS_MESSAGE_DELETE_FAIL] = "删除班级消息失败，请稍后重试"

		errorMsgs[ERR_COURSE_MESSAGE_QUERY_FAIL] = "查询课程信息错误，请稍后重试"
		errorMsgs[ERR_COURSE_DOWNLOAD_FAIL] = "课程下载失败，请稍后重试"
		errorMsgs[ERR_COURSE_BROWSE_FAIL] = "记录浏览量错误"
		errorMsgs[ERR_LESSIONS_MESSAGE_QUERY_FAIL] = "查询指定课程下的课时列表信息错误，请稍后重试"

		errorMsgs[ERR_LESSION_PROGRESS_UPDATE_FAIL] = "更新课节进展错误，请稍后重试"
		errorMsgs[ERR_LESSION_STATUS_UPDATE_FAIL] = "课节已学完，状态无法修改"
		errorMsgs[ERR_LESSION_PROGRESS_QUERY_FAIL] = "课节学习状态查询错误，请稍后重试"

		errorMsgs[ERR_TOOL_MESSAGE_QUERY_FAIL] = "查询工具信息错误，请稍后重试"
		errorMsgs[ERR_TOOL_DOWNLOAD_FAIL] = "工具下载失败,请稍后重试"

		errorMsgs[ERR_UPGRADE_LOGIN_TOKEN] = "无法登录升级服务器，请检查网络是否能正常访问互联网"
		errorMsgs[ERR_UPGRADE_CHECK_FAIL] = "请求升级检查失败，请检查网络是否能正常访问互联网"
		errorMsgs[ERR_UPGRADE_FAIL] = "升级失败"

		errorMsgs[ERR_UPGRADE_DOWNLOAD] = "平台升级下载失败"
		errorMsgs[ERR_UPGRADE_REBOOT] = "重启失败"
		errorMsgs[ERR_UPGRADE_DOEN] = "已经是最新版本"

		errorMsgs[ERR_NO_WORK_EXISTS] = "作品不存在"
		errorMsgs[ERR_NO_file_EXISTS] = "项目文件不存在"
		errorMsgs[ERR_SGL_NO_file_EXISTS] = "SGL项目文件不存在"
		errorMsgs[ERR_ADD_WORK_FAIL] = "作品保存失败"
		errorMsgs[ERR_LAUD_WORK_FAIL] = "作品点赞失败"
		errorMsgs[ERR_SHARE_WORK_FAIL] = "作品分享失败"
		errorMsgs[ERR_DELETE_WORK_FAIL] = "作品删除失败"
		errorMsgs[ERR_READ_WORK_FAIL] = "读取作品文件失败"
		errorMsgs[ERR_WRITE_WORK_FAIL] = "写入作品文件失败"
		errorMsgs[ERR_CREATE_FILE_FAIL] = "文件创建失败"
		errorMsgs[ERR_LAUD_RECORD_QUERY_FAIL] = "请勿重复点赞"
		errorMsgs[ERR_QUERY_FAVOR_FAIL] = "作品已收藏"
		errorMsgs[ERR_SELF_UNCOLLECTION_FAIL] = "无法收藏本人作品"
		errorMsgs[ERR_COLLECTION_WORK_FAIL] = "收藏作品失败"
		errorMsgs[ERR_FAVORITE_WORK_LIST] = "获取个人收藏失败"
		errorMsgs[ERR_DELETE_FAVORITE_WORK] = "删除个人收藏失败"
		errorMsgs[ERR_UPLOAD_IMAGES_FAIL] = "上传文件失败，请稍后重试"
		errorMsgs[ERR_COUNT_FAIL] = "统计总数错误"

		errorMsgs[ERR_CUSTOMECOURSE_EXIT] = "选课已存在"
		errorMsgs[ERR_CUSTOMECOURSE_NULL] = "暂时没有选课"
		errorMsgs[ERR_CUSTOMECOURSE_DELETE] = "删除课程错误"

		errorMsgs[ERR_DIRECTORY_CREATE] = "创建目录失败"

		errorMsgs[ERR_TOOL_REALPATH] = "无法找到作品对应的工具"

		errorMsgs[ERR_ADD_EXER_FAIL] = "练习保存失败"
		errorMsgs[RUNTIME_ERROR] = "系统错误，请稍后重试"

	}
	return errorMsgs
}

//GetErrorMsgs 根据错误码返回错误信息
func GetErrorMsgs(code int) string {
	msg, ok := ErrorMsgs()[code]
	if !ok {
		return "未知错误"
	}
	return msg
}
