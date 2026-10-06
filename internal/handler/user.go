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
	user, err := service.RegisterUser(a.StudentID, a.Password, a.Role)
	if err != nil {
		if errors.Is(err, service.ErrUserExists) {
			response.Fail(c, errcode.ErrResourceConflict)
			return
		} else if errors.Is(err, service.ErrStudentNotFound) {
			response.Fail(c, errcode.ErrStudentNotFound)
			return
		} else if errors.Is(err, service.ErrInvalidParams) {
			response.Fail(c, errcode.ErrInvalidParams)
			return
		} else {
			response.Fail(c, errcode.ErrServer)
			fmt.Println("❌ 注册失败:", err)
			return
		}
	}
	response.Success(c, gin.H{
		"student_id": user.StudentID,
		"name":       user.Name,
		"id":         user.ID,
		"role":       user.Role,
	})

}
func Login(c *gin.Context) {
	var b model.LoginReq
	err := c.ShouldBindJSON(&b)
	if err != nil {
		response.FailReason(c, errcode.ErrInvalidParams, err.Error())
		return
	}
	user1, err := service.LoginUser(b.StudentID, b.Password)
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
			"id":         user1.ID,
			"student_id": user1.StudentID,
			"role":       user1.Role,
			"name":       user1.Name,
			"avatar_url": user1.AvatarURL,
			"theme":      user1.Theme,
		},
	})
}

func GetCurrentUser(c *gin.Context) {
	userid := middleware.GetUserID(c)
	user, err := service.GetUserProfile(userid)
	if err != nil {
		response.Fail(c, errcode.ErrServer)
		return
	}
	response.Success(c, user)
	return

}

func Logout(c *gin.Context) {
	info := middleware.GetTokenInfo(c)
	if info == nil {
		response.Fail(c, errcode.ErrNoToken)
		return
	}
	err := service.SessionCancel(info.SID)
	if err != nil {
		response.Fail(c, errcode.ErrServer)
		fmt.Println("❌", err)
		return
	}
	response.SuccessMsg(c, "已退出登录", nil)
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
		if errors.Is(err, service.ErrInvalidParams) {
			response.Fail(c, errcode.ErrInvalidParams)
			return
		}
		response.Fail(c, errcode.ErrServer)
		fmt.Println("❌", err)
		return
	}
	response.SuccessMsg(c, "密码已修改，请重新登陆", nil)

}

func UpdateAvatar(c *gin.Context) {
	var req model.UpdateAvatarReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, errcode.ErrInvalidParams)
		return
	}
	user, err := service.UpdateAvatar(middleware.GetUserID(c), req.AvatarURL)
	if err != nil {
		if errors.Is(err, service.ErrInvalidParams) {
			response.Fail(c, errcode.ErrInvalidParams)
			return
		}
		response.Fail(c, errcode.ErrServer)
		return
	}
	response.Success(c, user)
}
