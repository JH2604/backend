package middleware

import (
	"errors"
	"gin-demo/internal/model"
	"gin-demo/internal/service"
	"gin-demo/pkg/errcode"
	"gin-demo/pkg/response"
	"strings"

	"github.com/gin-gonic/gin"
)

const key = "user_id"

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			response.Fail(c, errcode.ErrNoToken)
			c.Abort()
			return
		}
		token1 := header[7:]
		info, err := service.ValidateToken(token1)
		if err != nil {
			if errors.Is(err, service.ErrSessionInvalid) {
				response.Fail(c, errcode.ErrSessionInvalid) //40103会话没了，跳登录页
			} else {
				response.Fail(c, errcode.ErrNoToken)//40100 令牌本身不行
			}
			c.Abort()
			return
		}
		c.Set(key, info)
		c.Next()

	}
}

func GetUserID(c *gin.Context) uint {
	v, exist := c.Get(key)
	if !exist {
		return 0
	}
	id, ok := v.(*model.TokenInfo)
	if !ok {
		return 0
	}
	return id.UserID
}

func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exist := c.Get(key)
		if !exist {
			response.Fail(c, errcode.ErrNoToken)
			c.Abort()
			return
		}
		info, ok := role.(*model.TokenInfo)
		if !ok {
			response.Fail(c, errcode.ErrNoToken)
			c.Abort()
			return
		}
		for _, v := range roles {
			if info.Role == v {
				c.Next()
				return
			}

		}
		response.Fail(c, errcode.ErrPermission)
		c.Abort()
		return

	}
}

func GetTokenInfo(c *gin.Context) *model.TokenInfo {
	a, exist := c.Get(key)
	if !exist {
		return nil
	}
	id, ok := a.(*model.TokenInfo)
	if !ok {
		return nil
	}
	return id
}
