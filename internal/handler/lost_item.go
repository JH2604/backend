package handler

import (
	"gin-demo/internal/model"
	"gin-demo/internal/repository"
	"gin-demo/pkg/errcode"
	"gin-demo/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func CreateLostItem(c *gin.Context) {
	var item model.LostItem
	if err := c.ShouldBindJSON(&item); err != nil {
		response.FailReason(c, errcode.ErrInvalidParams, err.Error())
		return
	}
	if err := repository.CreateLostItem(&item); err != nil {
		response.Fail(c, errcode.ErrServer)
		return
	}
	response.Success(c, item)
}

func ListLostItems(c *gin.Context) {
	location := c.Query("location")
	items, err := repository.ListLostItems(location)
	if err != nil {
		response.Fail(c, errcode.ErrServer)
		return
	}
	response.Success(c, items)
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
