package handler

import (
	"github.com/gin-gonic/gin"

	"campus-lost-found/backend/models"
	"campus-lost-found/backend/response"
)

func GetItem(c *gin.Context) {

	item := models.Item{ID: 1, Name: "黑色雨伞", Type: "lost", Status: "pending"}
	response.Success(c, item)

}
