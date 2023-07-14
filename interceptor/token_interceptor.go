package interceptor

import (
	"FanCode/result"
	"FanCode/setting"
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
		for _, releaseStartPath := range setting.Conf.ReleasePathConfig.StartWith {
			if strings.HasPrefix(path, releaseStartPath) {
				c.Next()
				return
			}
		}
		// 检验是否携带token
		r := result.NewResult(c)
		token := c.Request.Header.Get("token")
		user, err := utils.ParseToken(token)
		if err != nil || user == nil {
			r.Error(result.IDENTITY_INVALID.GetCode(),
				result.IDENTITY_INVALID.GetMessage(), nil)
			return
		}
		if c.Keys == nil {
			c.Keys = make(map[string]interface{}, 1)
		}
		c.Keys["user"] = user
	}
}
