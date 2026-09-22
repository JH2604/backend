package middleware

import (
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
			response.Fail(c, errcode.ErrUnauthorized)
			c.Abort()
			return
		}
		token1 := header[7:]
		userID, err := service.ParseToken(token1)
		if err != nil {
			response.Fail(c, errcode.ErrUnauthorized)
			c.Abort()
			return
		}
		c.Set(key, userID)
		c.Next()

	}
}

func GetUserID(c *gin.Context)uint{
	v,exist:=c.Get(key)
	if !exist{
		return 0
	}
	id,ok:=v.(uint)
	if !ok{
		return 0
	}
	return id
}
