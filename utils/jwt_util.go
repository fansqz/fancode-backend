package utils

import (
	"github.com/dgrijalva/jwt-go"
	"time"
)

const (
	key         = "fan_code_key...naliyoucaihonggaosuwo"
	expiredTime = 12 * time.Hour
)

type Claims struct {
	ID        uint     `json:"id"`
	Username  string   `json:"username"`
	LoginName string   `json:"loginName"`
	Phone     string   `json:"phone"`
	Email     string   `json:"email"`
	Roles     []uint   `json:"roles"`
	Menus     []string `json:"menus"`
	jwt.StandardClaims
}

// GenerateToken
// https://www.jianshu.com/p/202b04426368
//
//	@Description:  通过user生成token
//	@param user    用户
//	@return string 生成的token
//	@return error
func GenerateToken(claims Claims) (string, error) {
	nowTime := time.Now()
	expiredTime := nowTime.Add(expiredTime)
	claims.StandardClaims = jwt.StandardClaims{
		//过期时间
		ExpiresAt: expiredTime.Unix(),
		//指定token发行人
		Issuer: "fancode",
	}
	//设置加密算法，生成token对象
	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	//通过私钥获取已签名token
	token, err := tokenClaims.SignedString([]byte(key))
	return token, err
}

// ParseToken
//
//	@Description: 解析token，返回user
//	@param token
//	@return user
func ParseToken(token string) (*Claims, error) {
	//获取到token对象
	tokenClaims, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(key), nil
	})
	//通过断言获取到claim
	if tokenClaims != nil {
		if claims, ok := tokenClaims.Claims.(*Claims); ok && tokenClaims.Valid {
			return claims, nil
		}
	}
	return nil, err
}
