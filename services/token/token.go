// @APIVersion 1.0.0
// @Title 通行令牌服务
// @Description 采用JWT生成用户登录通行令牌TOKEN
// @Contact xuchuangxin@icanmake.cn
// @TermsOfServiceUrl https://maiyajia.com/
// @License
// @LicenseUrl

package token

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/astaxie/beego"
	jwt "github.com/dgrijalva/jwt-go"
)

var (
	tokenSecretKey  string
	tokenIssuer     string
	tokenExpireTime time.Duration
	timeLimit       time.Duration
)

// Token 自定义token
type Token struct {
	UserID      string
	Username    string
	UserRole    string
	CreateTime  int64
	ExpireTime  int64
	TokenIssuer string
}

// EmailToken 自定义邮箱令牌
type EmailToken struct {
	UserID      string
	Email       string
	CreateTime  int64
	ExpireTime  int64
	TokenIssuer string
}

func init() {
	tokenSecretKey = beego.AppConfig.String("token_secret_key")
	tokenIssuer = beego.AppConfig.DefaultString("token_issuer", "malacode.com")

	hour := beego.AppConfig.DefaultInt("token_expire_time", 1)
	tokenExpireTime = time.Hour * time.Duration(24*hour)
	minute := beego.AppConfig.DefaultInt("email_expire_time ", 10)
	timeLimit = time.Minute * time.Duration(minute)
}

/** CreateEmailToken 创建JWT的TOKEN签名字符串
* 返回: 创建成功则返回token字符串，失败则返回错误信息
*[参见文档:https://godoc.org/github.com/dgrijalva/jwt-go#example-New--Hmac]
 */
func (et *EmailToken) CreateEmailToken() (string, error) {
	timestamp := time.Now()

	if et.CreateTime == 0 {
		et.CreateTime = timestamp.Unix()
	}
	if et.ExpireTime == 0 {
		et.ExpireTime = timestamp.Add(timeLimit).Unix()
	}
	if len(et.TokenIssuer) == 0 {
		et.TokenIssuer = tokenIssuer
	}
	claims := &jwt.StandardClaims{
		Audience:  et.UserID,
		Subject:   et.Email,
		Issuer:    et.TokenIssuer,
		IssuedAt:  et.CreateTime,
		ExpiresAt: et.ExpireTime,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, signErr := token.SignedString([]byte(tokenSecretKey))

	if signErr != nil {
		log.Fatal(signErr)
	}

	return tokenString, signErr
}

/** CreateToken 创建JWT的TOKEN签名字符串
* 返回: 创建成功则返回token字符串，失败则返回错误信息
*[参见文档:https://godoc.org/github.com/dgrijalva/jwt-go#example-New--Hmac]
 */
func (t *Token) CreateToken() (string, error) {

	timestamp := time.Now()

	if t.CreateTime == 0 {
		t.CreateTime = timestamp.Unix()
	}
	if t.ExpireTime == 0 {
		t.ExpireTime = timestamp.Add(tokenExpireTime).Unix()
	}
	if len(t.TokenIssuer) == 0 {
		t.TokenIssuer = tokenIssuer
	}

	/** cliams标准字段的说明 [参考文档 http://blog.zhishile.com/Article/Show/90e3cb1f-bbe5-4dbc-97f2-7ce7c8e83593]
	Cliams 也有称之为Payloads，是存放有效信息的地方。这个部分包含了业务中的所有的信息，用户可以任意自定义，但是要避免过于复杂或者太多的Claims影响性能。
	标准中注册的声明 (建议但不强制使用) ：
	iss: jwt签发者
	sub: jwt所面向的用户
	aud: 接收jwt的一方
	exp: jwt的过期时间，这个过期时间必须要大于签发时间
	nbf: 定义在什么时间之前，该jwt都是不可用的.
	iat: jwt的签发时间
	jti: jwt的唯一身份标识，主要用来作为一次性token,从而回避重放攻击。
	*/
	claims := &jwt.StandardClaims{
		Id:        t.UserID,
		Audience:  t.UserID,
		Subject:   t.Username + "/" + t.UserRole, // 注意: 这里把用户名和用户角色通过"/"合并为一个字符串信息
		Issuer:    t.TokenIssuer,
		IssuedAt:  t.CreateTime,
		ExpiresAt: t.ExpireTime,
		Username:  t.Username,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, signErr := token.SignedString([]byte(tokenSecretKey))

	if signErr != nil {
		log.Fatal(signErr)
	}

	return tokenString, signErr
}

// ParseEmailToken 验证并解析邮箱令牌
func ParseEmailToken(emailToken string) (*EmailToken, error) {

	// 验证token，返回token的claims
	token, err := ValidatesToken(emailToken)
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*jwt.StandardClaims)
	if !ok {
		return nil, errors.New("Parse emailToken is invalid")
	}
	return &EmailToken{
		UserID:      claims.Audience,
		Email:       claims.Subject,
		TokenIssuer: claims.Issuer,
		CreateTime:  claims.IssuedAt,
		ExpireTime:  claims.ExpiresAt,
	}, nil

}

/** ParseToken 验证并解析TOKEN签名字符串
* 参数: TOKEN签名字符串
** 返回: 创建成功则返回Token{}，失败则返回错误信息
 */
func ParseToken(tokenString string) (*Token, error) {

	// 验证token，返回token的claims
	token, err := ValidatesToken(tokenString)
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*jwt.StandardClaims)
	if !ok {
		return nil, errors.New("Claims is invalid")
	}
	// 从Subject中分解出UserName和UserRole
	subs := strings.Split(claims.Subject, "/")
	if len(subs) != 2 {
		return nil, errors.New("Claims is invalid")
	}
	return &Token{
		UserID:      claims.Audience,
		Username:    subs[0],
		UserRole:    subs[1],
		TokenIssuer: claims.Issuer,
		CreateTime:  claims.IssuedAt,
		ExpireTime:  claims.ExpiresAt,
	}, nil

}

// ValidatesToken 验证token字符串是否有效
func ValidatesToken(tokenString string) (*jwt.Token, error) {
	if len(tokenString) == 0 {
		return nil, errors.New("token is empty")
	}

	token, err := jwt.ParseWithClaims(tokenString, &jwt.StandardClaims{}, func(jt *jwt.Token) (interface{}, error) {
		// 提示: SigningMethodHS256加密对应的就是SigningMethodHMAC
		if _, ok := jt.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", jt.Header["alg"])
		}
		return []byte(tokenSecretKey), nil
	})

	if token != nil && token.Valid {
		return token, nil

	} else if valid, ok := err.(*jwt.ValidationError); ok {
		// 解析错误，然后返回
		if valid.Errors&jwt.ValidationErrorMalformed != 0 {
			return nil, errors.New("This's not a token")
		} else if valid.Errors&(jwt.ValidationErrorExpired|jwt.ValidationErrorNotValidYet) != 0 {
			return nil, errors.New("Token is either expired or not active yet")
		} else {
			return nil, err
		}
	} else {
		return nil, errors.New("Couldn't handle this token")
	}

}
