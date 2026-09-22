// 数据：GORM 存
package repository

import (
	"gin-demo/internal/model"
)

func CreateUser(v *model.User) error {
	return db.Create(v).Error

}
func FindByUsername(username string) (model.User, error) {
	var user model.User
	err := db.Where("username = ?", username).First(&user).Error
	if err != nil {
		return model.User{}, err
	}
	return user, nil

}
