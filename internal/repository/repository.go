package repository

// GORM 常用写法速查（平话版）：docs/gorm.md

import (
	"gorm.io/gorm"
)

var db *gorm.DB

func Init(c *gorm.DB) {
	db = c
}
