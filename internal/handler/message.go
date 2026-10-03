package handler

import (
	"fmt"
	"gin-demo/internal/middleware"
	"gin-demo/internal/service"
	"gin-demo/pkg/errcode"
	"gin-demo/pkg/response"

	"github.com/gin-gonic/gin"
)

func GetUnreadCount(c *gin.Context) {
	userID := middleware.GetUserID(c)

	count, err := service.GetUnreadCount(userID)
	if err != nil {
		response.Fail(c, errcode.ErrServer)
		fmt.Println("❌ 查询未读数失败:", err)
		return
	}
	response.Success(c, gin.H{"unread": count})

}
