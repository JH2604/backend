package repository

import (
	"gin-demo/internal/model"

	"gorm.io/gorm"
)

var db *gorm.DB

func Init(c *gorm.DB) {
	db = c
}

func CreateUser(v *model.User) error {
	return db.Create(v).Error

}
