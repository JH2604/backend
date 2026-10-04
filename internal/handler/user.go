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
			response.Fail(c, errcode.ErrResourceConflict)
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
			response.Fail(c, errcode.ErrBadCredentials)
			return
		}
		response.Fail(c, errcode.ErrServer)
		fmt.Println("❌服务器内部错误:", err)
		return
	}
	accessToken, refreshToken, err := service.GenerateTokenPair(
		user1,
		c.GetHeader("X-Client-Platform"),
		c.ClientIP(),
		c.Request.UserAgent(),
	)
	if err != nil {

		response.Fail(c, errcode.ErrServer)
		fmt.Println("❌", err)
		return

	}

	response.Success(c, gin.H{
		"access_token":       accessToken,
		"refresh_token":      refreshToken,
		"token_type":         "Bearer",
		"expires_in":         int(service.AccessTokenTTL.Seconds()),
		"refresh_expires_in": int(service.RefreshTokenTTL.Seconds()),
		"user": gin.H{
			"id":       user1.ID,
			"username": user1.Username,
			"role":     user1.Role,
		},
	})
}

func GetCurrentUser(c *gin.Context) {
	info := middleware.GetTokenInfo(c)
	response.Success(c, gin.H{"user_id": info.UserID, "username": info.Username, "role": info.Role})

}

func Logout(c *gin.Context) {
	info := middleware.GetTokenInfo(c)
	err := service.SessionCancel(info.SID)
	if err != nil {
		response.Fail(c, errcode.ErrServer)
		fmt.Println("❌", err)
		return
	}
	response.Success(c, "注销成功")
}

func Refresh(c *gin.Context) {
	var req model.RefreshReq
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.FailReason(c, errcode.ErrInvalidParams, err.Error())
		return
	}

	access, refresh, err := service.RefreshTokens(req.RefreshToken)
	if err != nil {
		if errors.Is(err, service.ErrRefreshTokenInvalid) {
			response.Fail(c, errcode.ErrRefreshToken)
			return
		}
		response.Fail(c, errcode.ErrServer)
		fmt.Println("❌", err)
		return
	}

	response.Success(c, gin.H{
		"access_token":       access,
		"refresh_token":      refresh,
		"token_type":         "Bearer",
		"expires_in":         int(service.AccessTokenTTL.Seconds()),
		"refresh_expires_in": int(service.RefreshTokenTTL.Seconds()),
	})
}

// 修改密码
func ChangePassword(c *gin.Context) {
	var user model.PasswordChange
	err := c.ShouldBindJSON(&user)
	if err != nil {
		response.Fail(c, errcode.ErrInvalidParams)
		return
	}
	userID := middleware.GetUserID(c)
	err = service.ChangePassword(userID, user.OldPassword, user.NewPassword)
	if err != nil {
		if errors.Is(err, service.ErrOldPasswordWrong) {
			response.Fail(c, errcode.ErrOldPassword)
			return
		}
		response.Fail(c, errcode.ErrServer)
		fmt.Println("❌", err)
		return
	}
	response.SuccessMsg(c, "密码已修改，请重新登陆", nil)

}
