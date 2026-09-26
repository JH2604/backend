// HTTP 的事
package handler

import (
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"

	"gin-demo/internal/middleware"
	"gin-demo/internal/model"
	"gin-demo/internal/service"
	"gin-demo/pkg/errcode"
	"gin-demo/pkg/response"
)

func Register(c *gin.Context) {
	var a model.RegisterReq
	err := c.ShouldBindJSON(&a)
	if err != nil {
		response.FailReason(c, errcode.ErrInvalidParams, err.Error())
		return
	}
	user, err := service.RegisterUser(a.Username, a.Password)
	if err != nil {
		if errors.Is(err, service.ErrUserExists) {
			response.Fail(c, errcode.ErrUserExists)
			return
		}
		response.Fail(c, errcode.ErrServer)
		fmt.Println("❌ 注册失败:", err)
		return
	}
	response.Success(c, user.Username)

}
func Login(c *gin.Context) {
	var b model.LoginReq
	err := c.ShouldBindJSON(&b)
	if err != nil {
		response.FailReason(c, errcode.ErrInvalidParams, err.Error())
		return
	}
	user1, err := service.LoginUser(b.Username, b.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			response.Fail(c, errcode.ErrUserFormat)
			return
		}
		response.Fail(c, errcode.ErrServer)
		fmt.Println("❌服务器内部错误:", err)
		return
	}
	createdtoken, err := service.CreateToken(user1)
	if err != nil {

		response.Fail(c, errcode.ErrServer)
		fmt.Println("❌", err)
		return

	}

	response.Success(c, gin.H{"token": createdtoken})
}

func GetCurrentUser(c *gin.Context) {
	info := middleware.GetTokenInfo(c)
	response.Success(c, gin.H{"user_id": info.UserID, "username": info.Username, "role": info.Role})

}
