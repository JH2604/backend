package repository

import (
	"gorm.io/gorm"
)

var db *gorm.DB

func Init(c *gorm.DB) {
	db = c
}
