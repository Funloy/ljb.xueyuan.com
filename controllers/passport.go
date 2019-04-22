// @APIVersion 1.0.0
// @Title 用户登录注册控制器
// @Description 本控制器提供的接口为整个平台提供用户注册、登录验证以及获取通行令牌TOKEN的功能。
// @Contact xuchuangxin@icanmake.cn
// @TermsOfServiceUrl https://maiyajia.com/
// @License
// @LicenseUrl

package controllers

import (
	"encoding/json"
	"regexp"
	"strings"
	"time"

	"github.com/astaxie/beego/logs"

	"github.com/astaxie/beego"

	"golang.org/x/crypto/bcrypt"
	"gopkg.in/mgo.v2/bson"

	"github.com/astaxie/beego/context"

	m "maiyajia.com/models"
	"maiyajia.com/services/daemon"
	"maiyajia.com/services/token"
)

var (
	domain          string
	tokenExpireTime time.Duration
)

// PassportController 登录注册控制器
type PassportController struct {
	BaseController
	userClient m.UserModels
	logMod     m.LoggerModels
}

// LoginCredential 登录凭证
type loginCredential struct {
	Username string `form:"username"`
	Password string `form:"password"`
}

// RegisterCredential 注册凭证
// 用户名，密码，真实姓名，年级，班级
type registerCredential struct {
	Username string `form:"username"`
	Password string `form:"password"`
	Role     string `form:"role"`
	ProdKey  string `form:"prodKey"`
	RealName string `form:"realname"`
	Gender   int    `form:"gender"`
	Grade    int    `form:"grade"`
	Class    int    `form:"class"`
}

// PassCredential 用户修改密码凭证
type passCredential struct {
	OldPassword string `form:"oldPassword"`
	NewPassword string `form:"newPassword"`
}
type resetPass struct {
	ResetToken  string `form:"resettoken"`
	NewPassword string `form:"newPassword"`
}

// repassCredentail 教师为学生重置密码请求凭证
type repassStudentCredential struct {
	Username    string `form:"username"`
	NewPassword string `form:"newPassword"`
}

// repassCredentail 教师为学生重置密码请求凭证
type repassTeacherCredential struct {
	Username    string `form:"username"`
	NewPassword string `form:"newPassword"`
	Key         string `form:"key"`
}

//  blockedCredential 教师冻结/解封学生账户请求凭证
// blocked = true 表示冻结账号; blocked = false 表示解封账号
type blockedCredential struct {
	Username string `form:"username"`
	Blocked  bool   `form:"blocked"`
}

// NestPrepare 初始化函数
// 把控制器的MgoClient赋值到数据库操作客户端
func (passport *PassportController) NestPrepare() {
	passport.userClient.MgoSession = passport.MgoClient
	passport.logMod.MgoSession = &passport.MgoClient
}
func init() {
	domain = beego.AppConfig.String("domain")
	hour := beego.AppConfig.DefaultInt("token_expire_time", 1)
	tokenExpireTime = time.Hour * time.Duration(24*hour)
}

//ResetPass 重置密码
func (passport *PassportController) ResetPass() {
	//凭证解析
	var resetcredential resetPass
	if err := json.Unmarshal(passport.Ctx.Input.RequestBody, &resetcredential); err != nil {
		passport.abortWithError(m.ERR_REQUEST_PARAM)
	}
	resettoken, err := token.ParseEmailToken(resetcredential.ResetToken)
	if err != nil {
		passport.abortWithError(m.ERR_PARSE_REPASSTOKEN_FAIL)
	}
	// 重置用户密码
	hashPass, _ := hashPassword(resetcredential.NewPassword)
	if err := passport.userClient.ResetPassword(resettoken.UserID, hashPass); err != nil {
		passport.abortWithError(m.ERR_ADMIN_REPASS_FAIL)
	}
	// 封装返回数据
	out := make(map[string]interface{})
	out["code"] = 0
	// 返回结果
	passport.jsonResult(out)
}

//MailValid 验证邮件是否有效
func (passport *PassportController) MailValid() {
	var activateToken string
	if passport.Ctx.Input.Bind(&activateToken, "token") != nil {
		passport.abortWithError(m.ERR_REQUEST_PARAM)
	}
	logs.Info("activateToken:", activateToken)
	validtoken, err := token.ParseEmailToken(activateToken)
	if err != nil {
		passport.abortWithError(m.ERR_PARSE_REPASSTOKEN_FAIL)
	}
	// 封装返回数据
	out := make(map[string]interface{})
	out["code"] = 0
	out["email"] = validtoken.Email
	// 返回结果
	passport.jsonResult(out)
}

//RetrievePassMail 发送找回密码邮件
func (passport *PassportController) RetrievePassMail() {
	var email string
	if passport.Ctx.Input.Bind(&email, "account") != nil {
		passport.abortWithError(m.ERR_REQUEST_PARAM)
	}
	// 检查用户是否存在
	user, err := passport.userClient.FindUserByEmail(email)
	if err != nil {
		passport.abortWithError(m.ERR_EMAIL_NONE)
	}
	newRetriToken, err := createEmailToken(user.ID.Hex(), email)
	if err != nil {
		passport.abortWithError(m.ERR_RETRIEVE_TOKEN_FAIL)
	}
	go m.SendRetrieveEmail(newRetriToken, email)
	// if m.SendRetrieveEmail(newRetriToken, email) != nil {
	// 	passport.abortWithError(m.ERR_SEND_EMAIL_FAIL)
	// }
	// 封装返回数据
	out := make(map[string]interface{})
	out["code"] = 0
	// 返回结果
	passport.jsonResult(out)
}

//BindEmail 向邮箱发送绑定邮件
func (passport *PassportController) BindEmail() {
	var email string
	if passport.Ctx.Input.Bind(&email, "account") != nil {
		passport.abortWithError(m.ERR_REQUEST_PARAM)
	}
	_, err := passport.userClient.FindUserByEmail(email)
	if err == nil {
		passport.abortWithError(m.ERR_EMAIL_EXIST)
	}
	token := passport.checkToken()
	newActivateToken, err := createEmailToken(token.UserID, email)
	if err != nil {
		passport.abortWithError(m.ERR_ACTIVATE_TOKEN_FAIL)
	}
	go m.SendActivationEmail(newActivateToken, email, token.Username)
	// if m.SendActivationEmail(newActivateToken, email, token.Username) != nil {
	// 	passport.abortWithError(m.ERR_SEND_EMAIL_FAIL)
	// }
	// 封装返回数据
	out := make(map[string]interface{})
	out["code"] = 0
	passport.jsonResult(out)
}

//ActivateEmail 激活邮箱
func (passport *PassportController) ActivateEmail() {
	var activateToken string
	if passport.Ctx.Input.Bind(&activateToken, "token") != nil {
		passport.abortWithError(m.ERR_REQUEST_PARAM)
	}
	logs.Info("activateToken:", activateToken)
	activetoken, err := token.ParseEmailToken(activateToken)
	if err != nil {
		passport.abortWithError(m.ERR_PARSE_ACTIVATETOKEN_FAIL)
		// code := m.ERR_PARSE_ACTIVATETOKEN_FAIL
		// result := m.ErrorResult{
		// 	Code:    code,
		// 	Message: m.GetErrorMsgs(code),
		// }
		// path := beego.AppConfig.String("activate_fail_redirect_url")
		// fail_redirect := fmt.Sprintf(path, result)
		// passport.Ctx.Redirect(302, fail_redirect)
	}
	if err := passport.userClient.UpsertEmail(activetoken); err != nil {
		passport.abortWithError(m.ERR_BIND_EMAIL_FAIL)
	}
	logs.Info("邮箱绑定成功")
	// 封装返回数据
	out := make(map[string]interface{})
	out["code"] = 0
	// 返回结果
	passport.jsonResult(out)
}

// @Title 用户登录
// @Description 用户登录
// @Param	username	body 	string	true	"用户名"
// @Param	password	body 	string	true	"密码"
// @Success 200 {object}
// @Failure 400 参数错误：缺失或格式错误
func (passport *PassportController) Login() {

	var credential loginCredential

	// 登录凭证解析错误
	if err := json.Unmarshal(passport.Ctx.Input.RequestBody, &credential); err != nil {
		passport.abortWithError(m.ERR_REQUEST_PARAM)
	}

	// 根据用户名，从数据库获取用户信息, 检查账号是否被锁住
	user := validateAccount(passport, credential.Username)

	// 对比密码
	if err := safeComparePassword(user.Password, []byte(credential.Password)); err != nil {
		passport.abortWithError(m.ERR_LOGIN_PWD_FAIL)
	}

	// 登录成功，用UserID和username来获取JWT的TOKEN
	token, err := createClientToken(user.ID.Hex(), user.Username, user.Role.Name)
	if err != nil {
		passport.abortWithError(m.ERR_LOGIN_TOKEN_FAIL)
	}
	if err := passport.userClient.InsertUserRecord(user); err != nil {
		logs.Info("登陆用户记录失败")
	}
	passport.Ctx.SetCookie("token", token, tokenExpireTime, "/", domain, false, true)
	// 用户登录日志
	passport.logging(passport.Ctx, user.ID)

	// 封装返回数据
	out := make(map[string]interface{})
	out["code"] = 0
	out["token"] = token
	out["user"] = user
	// 返回结果
	passport.jsonResult(out)
}

// @Title 用户注册
// @Description 用户注册，根据role参数的不同，注册为教师或学生账户
// @Param	username	body 	string	true	"用户名"
// @Param	password	body 	string	true	"密码"
// @Param	role		body	string	true    "用户类型（student代表学生，teacher代表教师，其他字符串均为不合法）"
// @Param	realname	body 	string	true 	"真实姓名"
// @Param	gender		body 	int		true	"性别， 0代表男，1代表女"
// @Param	grade		body 	int		true|false 	"年级（1~12）"
// @Param	class		body 	int		true|false	"班级"
// @Success 0 {object}
// @Failure 非0 参数错误：缺失或格式错误
func (passport *PassportController) Register() {

	var credential registerCredential

	// 验证用户的注册信息: 验证注册请求是否有效
	if err := json.Unmarshal(passport.Ctx.Input.RequestBody, &credential); err != nil {
		passport.abortWithError(m.ERR_REQUEST_PARAM)
	}
	// 验证用户的注册信息
	if !validateUsername(credential.Username) {
		passport.abortWithError(m.ERR_USER_FMT_FAIL)
	}
	if !validatePassword(credential.Password) {
		passport.abortWithError(m.ERR_PWD_FMT_FAIL)
	}

	//检测用户名是否被占用
	if _, err := passport.userClient.FindUserByUsername(credential.Username); err == nil {
		passport.abortWithError(m.ERR_USER_EXISTS)
	}

	//加密密码
	hashPass, _ := hashPassword(credential.Password)

	var user interface{}
	// 检查用户注册的类型是学生还是老师
	if credential.Role == m.ROLE_STUDENT {
		//创建新的学生用户
		user = m.NewStudent(credential.Username, hashPass,
			credential.RealName, credential.Gender, credential.Grade, credential.Class)

	} else if credential.Role == m.ROLE_TEACHER {
		// 注册教师需要提供产品的KEY，检查KEY是否是符合要求
		productKey, _, _ := daemon.GetProductInfo()
		if credential.ProdKey != productKey {
			passport.abortWithError(m.ERR_REGISTER_KEY_FAIL)
		}
		//创建新的教师用户
		user = m.NewTeacher(credential.Username, hashPass, credential.RealName, credential.Gender)
	} else {
		passport.abortWithError(m.ERR_REGISTER_FAIL)
	}

	err := passport.userClient.RegisteredUser(&user)
	if err != nil {
		passport.abortWithError(m.ERR_REGISTER_FAIL)
	}

	// 封装返回数据
	out := make(map[string]interface{})
	out["code"] = 0
	out["user"] = user
	// 返回结果
	passport.jsonResult(out)

}

// @Title 修改密码
// @Description 修改用户密码
// @Param	password	body 	string	true	"用户新密码"
// @Param	repassword	 	body	int		true	"新密码确认"
// @Success 200 {object}
// @Failure 400 参数错误：缺失或格式错误
func (passport *PassportController) ChangePass() {

	// 获取token
	token := passport.checkToken()

	// 获取凭证
	var credential passCredential
	if err := json.Unmarshal(passport.Ctx.Input.RequestBody, &credential); err != nil {
		passport.abortWithError(m.ERR_REQUEST_PARAM)
	}
	// 验证新密码是否符合要求
	if !validatePassword(credential.NewPassword) {
		passport.abortWithError(m.ERR_PWD_FMT_FAIL)
	}

	// 检查账号是否被锁住
	user := validateAccount(passport, token.Username)

	// 验证用户的原密码是否正确
	if err := safeComparePassword(user.Password, []byte(credential.OldPassword)); err != nil {
		passport.abortWithError(m.ERR_PWD_CHAN_OLD_FAIL)
	}

	// 更新用户密码
	hashPass, _ := hashPassword(credential.NewPassword)
	if err := passport.userClient.UpdatePassword(token.Username, hashPass); err != nil {
		passport.abortWithError(m.ERR_PWD_CHAN_UPDATE_FAIL)
	}

	// 密码变更后，更新用户的登录令牌(token)
	newToken, err := createClientToken(user.ID.Hex(), user.Username, user.Role.Name)
	if err != nil {
		passport.abortWithError(m.ERR_LOGIN_TOKEN_FAIL)
	}
	// 用户修改密码日志
	passport.logging(passport.Ctx, bson.ObjectIdHex(token.UserID))

	// 封装返回数据
	out := make(map[string]interface{})
	out["code"] = 0
	out["token"] = newToken
	// 返回结果
	passport.jsonResult(out)

}

/*********************************************************************************************/
/*********************************** 以下为教师管理学生账户的接口 *********************************/
/*********************************** *********************************************************/

// ResetStudentPass 教师为学生用户重置密码
func (passport *PassportController) ResetStudentPass() {
	// 获取token
	token := passport.checkToken()
	// 检查权限（角色为老师或管理员）
	passport.needAdminOrTeacherPermission(token)

	// 凭证解析错误
	var credential repassStudentCredential
	if err := json.Unmarshal(passport.Ctx.Input.RequestBody, &credential); err != nil {
		passport.abortWithError(m.ERR_REQUEST_PARAM)
	}
	// 验证新密码是否符合要求
	if !validatePassword(credential.NewPassword) {
		passport.abortWithError(m.ERR_PWD_FMT_FAIL)
	}

	// 检查账号是否被锁住
	user := validateAccount(passport, credential.Username)

	// 检查被修改的用户是否是学生，不是学生则不能为其重置密码
	if user.Role.Name != m.ROLE_STUDENT {
		passport.abortWithError(m.ERR_ADMIN_PERMISSION_ROLE)
	}

	// 重置用户密码
	hashPass, _ := hashPassword(credential.NewPassword)
	if err := passport.userClient.UpdatePassword(user.Username, hashPass); err != nil {
		passport.abortWithError(m.ERR_ADMIN_REPASS_FAIL)
	}

	// 密码重置日志
	passport.logging(passport.Ctx, bson.ObjectIdHex(token.UserID))

	// 封装返回数据
	out := make(map[string]interface{})
	out["code"] = 0
	// 返回结果
	passport.jsonResult(out)

}

// ResetTeacherPass 教师为自己重置密码，需要用到产品的key
func (passport *PassportController) ResetTeacherPass() {
	// 获取token
	token := passport.checkToken()
	// 检查权限
	passport.needAdminOrTeacherPermission(token)

	// 凭证解析错误
	var credential repassTeacherCredential
	if err := json.Unmarshal(passport.Ctx.Input.RequestBody, &credential); err != nil {
		passport.abortWithError(m.ERR_REQUEST_PARAM)
	}
	// 验证新密码是否符合要求
	if !validatePassword(credential.NewPassword) {
		passport.abortWithError(m.ERR_PWD_FMT_FAIL)
	}

	// 验证产品的Key是否正确
	productKey, _, _ := daemon.GetProductInfo()
	if credential.Key != productKey {
		passport.abortWithError(m.ERR_REGISTER_KEY_FAIL)
	}

	// 检查账号是否有效
	user := validateAccount(passport, credential.Username)
	// 检查被修改的用户是否是教师，不是教师则不能为通过这个接口重置账号密码
	if user.Role.Name != m.ROLE_TEACHER {
		passport.abortWithError(m.ERR_ADMIN_PERMISSION_ROLE)
	}

	// 重置用户密码
	hashPass, _ := hashPassword(credential.NewPassword)
	if err := passport.userClient.UpdatePassword(user.Username, hashPass); err != nil {
		passport.abortWithError(m.ERR_ADMIN_REPASS_FAIL)
	}

	// 密码重置日志
	passport.logging(passport.Ctx, bson.ObjectIdHex(token.UserID))

	// 封装返回数据
	out := make(map[string]interface{})
	out["code"] = 0
	// 返回结果
	passport.jsonResult(out)

}

// BlockedAccount 教师冻结学生账户
func (passport *PassportController) BlockedAccount() {
	// 获取token
	token := passport.checkToken()
	// 检查权限
	passport.needAdminOrTeacherPermission(token)

	// 凭证解析错误
	var credential blockedCredential
	if err := json.Unmarshal(passport.Ctx.Input.RequestBody, &credential); err != nil {
		passport.abortWithError(m.ERR_REQUEST_PARAM)
	}

	user, _ := passport.userClient.FindUserByUsername(credential.Username)
	if len(user.ID) == 0 {
		passport.abortWithError(m.ERR_USER_NONE)
	}

	// 检查被修改的用户是否是学生，不是学生则不能冻结/解冻账号
	if user.Role.Name != m.ROLE_STUDENT {
		passport.abortWithError(m.ERR_ADMIN_PERMISSION_ROLE)
	}

	if user.Locked == credential.Blocked {
		if credential.Blocked {
			passport.abortWithError(m.ERR_ADMIN_BLOCKED_EXISTS)
		} else {
			passport.abortWithError(m.ERR_ADMIN_UNBLOCKED_EXISTS)
		}
	}

	if err := passport.userClient.LockedUser(user.Username, credential.Blocked); err != nil {
		if credential.Blocked {
			passport.abortWithError(m.ERR_ADMIN_BLOCKED_FAIL)
		} else {
			passport.abortWithError(m.ERR_ADMIN_UNBLOCKED_FAIL)
		}
	}

	passport.logging(passport.Ctx, bson.ObjectIdHex(token.UserID))

	// 封装返回数据
	out := make(map[string]interface{})
	out["code"] = 0
	// 返回结果
	passport.jsonResult(out)

}

/*********************************************************************************************/
/*********************************** 以下为本控制器的辅助方法 ************************************/
/*********************************** *********************************************************/

// Log 记录用户访问日志
func (passport *PassportController) logging(context *context.Context, uid bson.ObjectId) {
	// 记录日志
	ip := context.Input.IP()
	client := context.Input.UserAgent()
	action := context.Input.URI()
	proxy := strings.Join(context.Input.Proxy(), ",")

	log := &m.PassportLog{
		UserID:      uid,
		LoginIP:     ip,
		LoginProxy:  proxy,
		LoginClient: client,
		Action:      action,
		LogTime:     time.Now(),
	}

	if err := passport.logMod.Logger(log); err != nil {
		beego.Error(err)
	}
}

// @Title 用户名验证
// @Description 验证用户名格式
// @Param	username	string	true	"用户名"
// @Success true {bool}
// @Failure false {bool}
func validateUsername(username string) bool {
	//用户名正则，4到20位（字母，数字，下划线，减号）
	reg := regexp.MustCompile(`^[a-zA-Z0-9_-]{4,20}$`)
	return reg.MatchString(username)

}

// @Title 密码验证
// @Description 验证用户密码格式
// @Param	password	string	true "密码"
// @Success true {bool}
// @Failure false {bool}
func validatePassword(password string) bool {
	if len(password) < 6 || len(password) > 16 {
		return false
	}
	return true
}

// validateAccount 根据用户名，从数据库获取用户信息, 检查账号是否有效（包括是否存在，是否被锁...)
func validateAccount(passport *PassportController, username string) *m.User {
	user, _ := passport.userClient.FindUserByUsername(username)
	if user == nil {
		passport.abortWithError(m.ERR_USER_NONE)
	}
	// 检查账号是否被锁住
	if user.Locked {
		passport.abortWithError(m.ERR_USER_LOCKED)
	}
	return user
}

/* HashPassword 密码加密
* 参数: 明文密码
* 返回: 密文密码和错误信息
 */
func hashPassword(password string) ([]byte, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	return hash, nil
}

/* SafeComparePassword 密码密文比较
* 参数: 两个
* 返回: 密码不相等，则返回错误信息
 */
func safeComparePassword(hash []byte, password []byte) error {
	err := bcrypt.CompareHashAndPassword(hash, password)
	return err
}

// @Title 创建登录令牌
// @Description 用通行证的信息创建用户登录的令牌TOKEN
// @Param	uid			string	true	"用户ID"
// @param	username	string	true	"用户名"
// @Success token {string}
// @Failure err {error}
func createClientToken(uid, username, userrole string) (string, error) {
	token := token.Token{
		UserID:   uid,
		Username: username,
		UserRole: userrole,
	}
	return token.CreateToken()
}

//createEmailToken 创建邮箱激活令牌
func createEmailToken(uid, email string) (string, error) {
	activateToken := token.EmailToken{
		UserID: uid,
		Email:  email,
	}
	return activateToken.CreateEmailToken()
}
