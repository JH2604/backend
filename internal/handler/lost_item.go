package handler

import (
	"fmt"
	"gin-demo/internal/middleware"
	"gin-demo/internal/model"
	"gin-demo/internal/repository"
	"gin-demo/internal/service"
	"gin-demo/pkg/errcode"
	"gin-demo/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func CreatePost(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req model.CreatePostReq
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.FailReason(c, errcode.ErrInvalidParams, err.Error())
		return
	}
	post, err := service.CreatePost(userID, req)
	if err != nil {
		response.Fail(c, errcode.ErrServer)
		fmt.Println("❌ 发帖失败:", err)
		return
	}
	response.SuccessCreated(c, post)
}

func ListLostItems(c *gin.Context) {
	location := c.Query("location")
	items, err := repository.ListLostItems(location)
	if err != nil {
		response.Fail(c, errcode.ErrServer)
		return
	}
	response.Success(c, items)
}

func GetLostItem(c *gin.Context) {
	id := c.Param("id")
	item, err := repository.GetLostItemByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Fail(c, errcode.ErrNotFound)
			return
		}
		response.Fail(c, errcode.ErrServer)
		return
	}
	response.Success(c, item)
}
