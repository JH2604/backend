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

func FindUserByID(id uint) (*model.User, error) {
	var user model.User
	err := db.Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func FindUsersByIDs(ids []uint)([]model.User,error){
	if len(ids) == 0{
		return nil,nil
	}
	var users []model.User
	err := db.Where("id IN ?",ids).Find(&users).Error
	if err != nil{
		return nil,err
	}
	return users,nil

}


// 存新密码函数
func SavePassword(userID uint,newpassword string) error {
	return db.Model(&model.User{}).Where("id=?", userID).Update("password",newpassword).Error
}