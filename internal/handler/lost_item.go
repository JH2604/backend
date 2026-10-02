package handler

import (
	"errors"
	"fmt"
	"gin-demo/internal/middleware"
	"gin-demo/internal/model"
	"gin-demo/internal/repository"
	"gin-demo/internal/service"
	"gin-demo/pkg/errcode"
	"gin-demo/pkg/response"
	"strconv"

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

func ListPosts(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var q model.ListPostsQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.FailReason(c, errcode.ErrInvalidParams, err.Error())
		return
	}
	posts, total, err := service.ListPosts(userID, &q)
	if err != nil {
		response.Fail(c, errcode.ErrServer)
		return
	}
	response.Success(c, response.Page{
		List:     posts,
		Total:    total,
		Page:     q.Page,
		PageSize: q.PageSize,
	})
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

func UpdatePostStatus(c *gin.Context) {
	n, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, errcode.ErrInvalidParams)
		return

	}
	userID := middleware.GetUserID(c)
	var req model.UpdatePostStatusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailReason(c, errcode.ErrInvalidParams, err.Error())
		return
	}

	post, err := service.UpdatePostStatus(uint(n), userID, req.Status)
	if err != nil {
		if errors.Is(err, service.ErrPostNotFound) {
			response.Fail(c, errcode.ErrNotFound)
		} else if errors.Is(err, service.ErrNotPostOwner) {
			response.Fail(c, errcode.ErrPermission)
		} else {
			response.Fail(c, errcode.ErrServer)
		}
		return
	}
	var msg string
	if post.Status == model.PostStatusOpen {
		msg = "已恢复为进行中"
	} else if post.Type == model.PostTypeLost {
		msg = "已标记为已找到"
	} else {
		msg = "已标记为已认领"
	}

	response.SuccessMsg(c, msg, gin.H{
		"id":        post.ID,
		"status":    post.Status,
		"closed_at": post.ClosedAt,
	})
}
