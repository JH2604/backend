package response

import (
	"github.com/gin-gonic/gin"

	"gin-demo/pkg/errcode"
)

type Response struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}

func Success(c *gin.Context, data any) {
	c.JSON(200, Response{
		Code: errcode.Success,
		Msg:  errcode.GetMsg(errcode.Success),
		Data: data,
	})
}

func Fail(c *gin.Context, code int) {
	c.JSON(200, Response{
		Code: code,
		Msg:  errcode.GetMsg(code),
		Data: nil,
	})
}

func FailReason(c *gin.Context, code int, reason string) {
	c.JSON(200, Response{
		Code: code,
		Msg:  errcode.GetMsg(code),
		Data: reason,
	})
}
