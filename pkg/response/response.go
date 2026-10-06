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

type Page struct {
	List     any   `json:"list"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
}

func writeJSON(c *gin.Context, httpStatus int, body Response) {
	if httpStatus == 401 {
		c.Header("WWW-Authenticate", `Bearer error="invalid_token"`)
	}
	c.JSON(httpStatus, body)
}

func Success(c *gin.Context, data any) {
	writeJSON(c, 200, Response{
		Code: errcode.Success,
		Msg:  errcode.GetMsg(errcode.Success),
		Data: data,
	})
}

func SuccessCreated(c *gin.Context, data any) {
	writeJSON(c, 201, Response{
		Code: errcode.Success,
		Msg:  errcode.GetMsg(errcode.Success),
		Data: data,
	})
}

func Fail(c *gin.Context, code int) {
	writeJSON(c, code/100, Response{
		Code: code,
		Msg:  errcode.GetMsg(code),
		Data: nil,
	})
}

func FailReason(c *gin.Context, code int, reason string) {
	writeJSON(c, code/100, Response{
		Code: code,
		Msg:  reason,
		Data: nil,
	})
}

func SuccessMsg(c *gin.Context, msg string, data any) {
	writeJSON(c, 200, Response{
		Msg:  msg,
		Data: data,
		Code: errcode.Success,
	})
}