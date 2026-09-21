package handler

import (
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"

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
