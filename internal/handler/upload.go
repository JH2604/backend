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
	usage := c.PostForm("usage")
	if usage == "" {
		response.Fail(c, errcode.ErrInvalidParams)
		return
	}
	if usage != "avatar" && usage != "post" {
		response.Fail(c, errcode.ErrInvalidParams)
		return
	}
	if err != nil {
		response.Fail(c, errcode.ErrInvalidParams)
		return
	}
	url, err := service.SavePhoto(fh, usage)
	if err != nil {
		if errors.Is(err, service.ErrPhotoTooLarge) {
			response.Fail(c, errcode.ErrFileTooLarge)
		} else if errors.Is(err, service.ErrPhotoType) {
			response.Fail(c, errcode.ErrFileType)
		} else {
			response.Fail(c, errcode.ErrServer)
		}
		return
	}

	response.SuccessCreated(c, url)

}
