package interceptor

import (
	"FanCode/constants"
	e "FanCode/error"
	"FanCode/models/dto"
	r "FanCode/models/vo"
	"FanCode/service"
	"FanCode/utils"
	"github.com/gin-gonic/gin"
	"strings"
)

// TokenAuthorize
//
//	@Description: token拦截器
//	@return gin.HandlerFunc
func TokenAuthorize(roleService service.SysRoleService, userService service.SysUserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检验是否携带token
		result := r.NewResult(c)
		// 读取token
		token := c.Request.Header.Get("token")
		var userInfo *dto.UserInfo
		if token != "" {
			claims, err2 := utils.ParseToken(token)
			userInfo = &dto.UserInfo{
				ID:        claims.ID,
				Avatar:    claims.Avatar,
				LoginName: claims.LoginName,
				Username:  claims.Username,
				Email:     claims.Email,
				Phone:     claims.Phone,
				Roles:     claims.Roles,
				Menus:     claims.Menus,
			}
			if err2 != nil || userInfo == nil {
				result.Error(e.ErrSessionInvalid)
				c.Abort()
				return
			}
			if c.Keys == nil {
				c.Keys = make(map[string]interface{}, 1)
			}
			c.Keys["user"] = userInfo
		}
		// 检验是否在放行名单
		apis, err := roleService.GetApisByRoleID(constants.TouristID)
		if err != nil {
			result.Error(e.ErrServer)
			c.Abort()
			return
		}
		path := c.Request.URL.Path
		method := c.Request.Method
		for _, api := range apis {
			if matchPath(path, api.Path) {
				if strings.EqualFold(method, constants.AllMethod) {
					c.Next()
					return
				} else if strings.EqualFold(method, api.Method) {
					c.Next()
					return
				}
			}
		}
		// 判断用户是否有权限访问该接口
		if userInfo == nil {
			c.Abort()
			return
		}
		rules, err := userService.GetRoleIDsByUserID(userInfo.ID)
		if err != nil {
			result.Error(err)
			c.Abort()
			return
		}
		for _, ruleID := range rules {
			apis, _ = roleService.GetApisByRoleID(ruleID)
			for _, api := range apis {
				if matchPath(path, api.Path) {
					if strings.EqualFold(method, constants.AllMethod) {
						c.Next()
						return
					} else if strings.EqualFold(method, api.Method) {
						c.Next()
						return
					}
				}
			}
		}
		rejectRequest(c)
	}
}

// 判断请求路径是否和规则相匹配
func matchPath(requestPath, pattern string) bool {
	routeSegments := strings.Split(requestPath, "/")
	patternSegments := strings.Split(pattern, "/")

	if len(routeSegments) != len(patternSegments) {
		return false
	}

	for i := 0; i < len(routeSegments); i++ {
		if patternSegments[i] != "" && patternSegments[i] != routeSegments[i] {
			if !strings.HasPrefix(patternSegments[i], ":") {
				return false
			}
		}
	}

	return true
}

func rejectRequest(ctx *gin.Context) {
	result := r.NewResult(ctx)
	result.Error(e.ErrPermissionInvalid)
	ctx.Abort()
}
