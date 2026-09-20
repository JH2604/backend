package handler

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"

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

	user := model.User{
		Username: a.Username,
		Password: a.Password,
	}
	err = repository.CreateUser(&user)
	if err != nil {
		reception := strings.Contains(err.Error(), "Duplicate entry")
		if reception == true {
			response.FailReason(c, errcode.Errcustomer, err.Error())
			return
		}
		response.Fail(c, errcode.ErrServer)
		fmt.Println("❌ 数据库连不上", err.Error())
		return

	}
	response.Success(c, a.Username)

}
