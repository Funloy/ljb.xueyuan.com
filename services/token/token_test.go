package token

import (
	"fmt"
	"testing"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

func TestCreateToken(t *testing.T) {

	tokenSecretKey = "malacode.com.token"
	tokenIssuer = "malacode.com"

	token := &Token{
		UserID:     "123456789",
		Username:   "18566208215",
		ExpireTime: time.Now().Unix() - 100212,
	}

	sign, _ := token.CreateToken()
	t.Logf("signature: %s", sign)

	tokenObj, parseErr := jwt.ParseWithClaims(sign, &jwt.StandardClaims{}, func(jt *jwt.Token) (interface{}, error) {
		// 提示: SigningMethodHS256加密对应的就是SigningMethodHMAC
		if _, ok := jt.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", jt.Header["alg"])
		}
		return []byte(tokenSecretKey), nil
	})

	fmt.Printf("%v\n", tokenObj.Valid)
	fmt.Printf("%v\n", parseErr)

	// k, err := ParseToken(sign)
	// if err != nil {
	// 	t.Errorf("token err: %v", err)
	// 	return
	// }
	// if k.ExpireTime > time.Now().Unix() {
	// 	t.Errorf("ExpireTime: %v", k.ExpireTime)
	// }

}
