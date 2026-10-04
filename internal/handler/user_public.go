package handler

import (
	"errors"
	"fmt"
	"strconv"

	"gin-demo/internal/middleware"
	"gin-demo/internal/service"
	"gin-demo/pkg/errcode"
	"gin-demo/pkg/response"

	"github.com/gin-gonic/gin"
)

// ListAdmins：U6 管理员列表（GET /users/admins）
func ListAdmins(c *gin.Context) {
	list, err := service.ListAdmins()
	if err != nil {
		response.Fail(c, errcode.ErrServer)
		fmt.Println("❌ 查询管理员列表失败:", err)
		return
	}
	// Apifox 里 data 直接是数组（data: [ {...} ]），不要再套 list/total
	response.Success(c, list)
}

// GetUserPublic：U7 查看发帖人信息（GET /users/{user_id}）
func GetUserPublic(c *gin.Context) {
	// ① 路径参数拿到的永远是字符串，先转成数字；
	//    转不动、或者传了 0，都算参数错误
	id, err := strconv.ParseUint(c.Param("user_id"), 10, 64)
	if err != nil || id == 0 {
		response.Fail(c, errcode.ErrInvalidParams)
		return
	}

	// ②「查看者是谁、什么角色」必须从 token 拿 ——
	//    这是「按角色裁剪字段」能成立的前提，前端说了不算
	viewer := middleware.GetTokenInfo(c)
	role := ""
	if viewer != nil {
		role = viewer.Role
	}

	user, err := service.GetUserPublic(uint(id), role)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			response.Fail(c, errcode.ErrNotFound)
			return
		}
		response.Fail(c, errcode.ErrServer)
		fmt.Println("❌ 查询用户信息失败:", err)
		return
	}
	response.Success(c, user)
}
