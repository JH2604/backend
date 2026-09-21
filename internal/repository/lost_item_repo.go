package repository

import (
	"gin-demo/internal/model"

	"gorm.io/gorm"
)

var DB *gorm.DB

func Init(db *gorm.DB) {
	DB = db
}

func CreateLostItem(item *model.LostItem) error {
	return DB.Create(item).Error
}

func ListLostItems(location string) ([]model.LostItem, error) {
	var items []model.LostItem
	query := DB.Model(&model.LostItem{})
	if location != "" {
		query = query.Where("location LIKE ?", "%"+location+"%")
	}
	err := query.Find(&items).Error
	return items, err
}

func GetLostItemByID(id string) (model.LostItem, error) {
	var item model.LostItem
	err := DB.First(&item, id).Error
	return item, err
}
