package handler

import (
	"errors"
	"gin-demo/internal/service"
	"gin-demo/pkg/errcode"
	"gin-demo/pkg/response"

	"github.com/gin-gonic/gin"
)

func UploadPhoto(c *gin.Context) {
	fh, err := c.FormFile("file")
	if err != nil {
		response.Fail(c, errcode.ErrInvalidParams)
		return
	}
	url, err := service.SavePhoto(fh)
	if err != nil {
		if errors.Is(err, service.ErrPhotoTooLarge) || errors.Is(err, service.ErrPhotoType) {
			response.FailReason(c, errcode.ErrInvalidParams, err.Error())
			return
		}
		response.Fail(c, errcode.ErrServer)
		return
	}
	response.Success(c, gin.H{"url": url})

}
