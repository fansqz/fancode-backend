package utils

import (
	"FanCode/models/dto"
	"github.com/dgrijalva/jwt-go"
	"time"
)

const (
	key         = "fan_code_key...naliyoucaihonggaosuwo"
	expiredTime = 12 * time.Hour
)

type claims struct {
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
func GenerateToken(user *dto.UserInfo) (string, error) {
	nowTime := time.Now()
	expiredTime := nowTime.Add(expiredTime)
	claims := claims{
		ID:        user.ID,
		Username:  user.Username,
		LoginName: user.LoginName,
		Phone:     user.Phone,
		Email:     user.Email,
		Menus:     user.Menus,
		Roles:     user.Roles,
		StandardClaims: jwt.StandardClaims{
			//过期时间
			ExpiresAt: expiredTime.Unix(),
			//指定token发行人
			Issuer: "fancode",
		},
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
func ParseToken(token string) (*dto.UserInfo, error) {
	//获取到token对象
	tokenClaims, err := jwt.ParseWithClaims(token, &claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(key), nil
	})
	//通过断言获取到claim
	if tokenClaims != nil {
		if claims, ok := tokenClaims.Claims.(*claims); ok && tokenClaims.Valid {
			user := &dto.UserInfo{}
			user.ID = claims.ID
			user.Username = claims.Username
			user.LoginName = claims.LoginName
			user.Phone = claims.Phone
			user.Email = claims.Email
			user.Roles = claims.Roles
			user.Menus = claims.Menus
			return user, nil
		}
	}
	return nil, err
}
