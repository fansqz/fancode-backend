package interceptor

import (
	e "FanCode/error"
	"FanCode/global"
	"FanCode/models/dto"
	result2 "FanCode/models/vo"
	"FanCode/utils"
	"github.com/gin-gonic/gin"
	"strings"
)

// TokenAuthorize
//
//	@Description: token拦截器
//	@return gin.HandlerFunc
func TokenAuthorize() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检验是否在放行名单
		path := c.Request.URL.Path
		for _, releaseStartPath := range global.Conf.ReleasePathConfig.StartWith {
			if strings.HasPrefix(path, releaseStartPath) {
				c.Next()
				return
			}
		}
		// 检验是否携带token
		r := result2.NewResult(c)
		token := c.Request.Header.Get("token")
		claims, err := utils.ParseToken(token)
		userInfo := &dto.UserInfo{
			ID:        claims.ID,
			LoginName: claims.LoginName,
			Username:  claims.Username,
			Email:     claims.Email,
			Phone:     claims.Phone,
			Roles:     claims.Roles,
			Menus:     claims.Menus,
		}
		if err != nil || userInfo == nil {
			r.Error(e.ErrSessionInvalid)
			c.Abort()
			return
		}
		if c.Keys == nil {
			c.Keys = make(map[string]interface{}, 1)
		}
		c.Keys["user"] = userInfo
	}
}
