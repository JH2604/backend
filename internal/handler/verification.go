package handler

import (
	"errors"
	"fmt"
	"gin-demo/internal/middleware"
	"gin-demo/internal/model"
	"gin-demo/internal/service"
	"gin-demo/pkg/errcode"
	"gin-demo/pkg/response"

	"github.com/gin-gonic/gin"
)

func SendVerificationCode(c *gin.Context) {
	var req model.SendCodeReq
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.FailReason(c, errcode.ErrInvalidParams, err.Error())

		return
	}
	if req.Channel == "sms" {
		response.Fail(c, errcode.ErrInvalidParams)
		fmt.Println("暂不支持短信")
		return
	}
	err = service.SendVerificationCode(req.Target, req.Scene)
	if err != nil {
		if errors.Is(err, service.ErrCodeTooFrequent) {
			response.Fail(c, errcode.ErrTooManyRequests)
			return
		}
		response.Fail(c, errcode.ErrServer)
		return
	}
	response.SuccessMsg(c, "验证码已发送", nil)

}

func BindContact(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req model.BindContactReq
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.FailReason(c, errcode.ErrInvalidParams, err.Error())
		return
	}
	if req.Channel == "sms" {
		response.Fail(c, errcode.ErrInvalidParams)
		fmt.Println("暂不支持短信")
		return
	}
	err = service.BindContact(userID, req.Target, "bind_contact", req.Code)
	if err != nil {
		if errors.Is(err, service.ErrCodeInvalid) {
			response.Fail(c, errcode.ErrCodeExpired)
			return
		} else if errors.Is(err, service.ErrContactTaken) {
			response.Fail(c, errcode.ErrResourceConflict)
			return
		} else {
			response.Fail(c, errcode.ErrServer)
		}
		return
	}
	response.SuccessMsg(c, "绑定成功", nil)

}
