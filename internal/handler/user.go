package handler

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"gin-demo/internal/model"
	"gin-demo/internal/repository"
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
	haxi, err := bcrypt.GenerateFromPassword([]byte(a.Password), bcrypt.DefaultCost)
	if err != nil {
		response.Fail(c, errcode.ErrServer)
		return
	}

	user := model.User{
		Username: a.Username,
		Password: string(haxi),
	}
	err = repository.CreateUser(&user)
	if err != nil {
		reception := strings.Contains(err.Error(), "Duplicate entry")
		if reception {
			response.FailReason(c, errcode.ErrUserExists, err.Error())
			return
		}
		response.Fail(c, errcode.ErrServer)
		fmt.Println("❌ 数据库操作失败", err.Error())
		return

	}
	response.Success(c, a.Username)

}
