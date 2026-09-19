package handler

import (
	"campus-lost-found/backend/errcode"
	"strconv"

	"github.com/gin-gonic/gin"

	"campus-lost-found/backend/models"
	"campus-lost-found/backend/response"
)

var items = []models.Item{
	{ID: 1, Name: "雨伞", Type: "lost", Status: "pending"},
	{ID: 2, Name: "手机", Type: "found", Status: "pending"},
	{ID: 3, Name: "钥匙", Type: "lost", Status: "pending"},
	{ID: 4, Name: "电脑", Type: "found", Status: "pending"},
}

func GetItem(c *gin.Context) {

	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		response.Fail(c, errcode.ErrInvalidParams)
		return
	}

	for _, v := range items {
		if id == v.ID {
			response.Success(c, v)
			return

		}
	}

	response.Fail(c, errcode.ErrNotFound)

}

func ListItems(c *gin.Context) {
	result, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil {
		response.Fail(c, errcode.ErrInvalidParams)
		return
	}
	if result <= 0 {
		response.Fail(c, errcode.ErrInvalidParams)
		return

	}
	extent, err := strconv.Atoi(c.DefaultQuery("size", "10"))
	if err != nil {
		response.Fail(c, errcode.ErrInvalidParams)
		return
	}

	if extent <= 0 {
		response.Fail(c, errcode.ErrInvalidParams)
		return
	}
	if extent > 50 {
		extent = 50
	}

	category := c.Query("type")
	news := make([]models.Item, 0)
	for _, v := range items {
		if category == v.Type || category == "" {

			news = append(news, v)

		}
	}
	if len(news) <= (result-1)*extent {
		new := make([]models.Item, 0)
		response.Success(c, new)
		return

	}
	end := result * extent // end是边界
	if end > len(news) {
		end = len(news)
	}

	slice := news[(result-1)*extent : end]

	response.Success(c, slice)
}
