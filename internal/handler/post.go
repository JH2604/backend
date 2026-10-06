package handler

import (
	"errors"
	"fmt"
	"gin-demo/internal/middleware"
	"gin-demo/internal/model"
	"gin-demo/internal/service"
	"gin-demo/pkg/errcode"
	"gin-demo/pkg/response"
	"io"
	"strconv"

	"github.com/gin-gonic/gin"
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
		if errors.Is(err, service.ErrInvalidParams) {
			response.Fail(c, errcode.ErrInvalidParams)
			return
		}
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

func GetPost(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, errcode.ErrInvalidParams)
		return
	}
	info := middleware.GetTokenInfo(c)
	if info == nil {
		response.Fail(c, errcode.ErrNoToken)
		return
	}
	post, err := service.GetPost(uint(id), info.UserID, info.Role)
	if err != nil {
		if errors.Is(err, service.ErrPostNotFound) {
			response.Fail(c, errcode.ErrNotFound)
			return
		} else {
			response.Fail(c, errcode.ErrServer)
		}
		return
	}
	response.Success(c, post)

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

func DeletePost(c *gin.Context) {
	n, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, errcode.ErrInvalidParams)
		return
	}
	info := middleware.GetTokenInfo(c)
	if info == nil {
		response.Fail(c, errcode.ErrNoToken)
		return
	}
	var req model.DeletePostReq
	if err := c.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		response.FailReason(c, errcode.ErrInvalidParams, err.Error())
		return
	}
	err = service.DeletePost(uint(n), info.UserID, info.Role, req.Reason)
	if err != nil {
		if errors.Is(err, service.ErrPostNotFound) {
			response.Fail(c, errcode.ErrNotFound)
		} else if errors.Is(err, service.ErrNoPermission) {
			response.Fail(c, errcode.ErrPermission)
		} else {
			response.Fail(c, errcode.ErrServer)
		}
		return
	}
	response.SuccessMsg(c, "删除成功", nil)

}
