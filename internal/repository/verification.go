package repository

import (
	"gin-demo/internal/model"
	"time"
)

func CreateVerificationCode(v *model.VerificationCode) error {
	return db.Create(v).Error
}

func FindLatestByTargetScene(target, scene string) (*model.VerificationCode, error) {
	var lastest model.VerificationCode
	err := db.Where("target = ?", target).Where("scene = ?", scene).Order("created_at DESC").First(&lastest).Error
	if err != nil {
		return nil, err
	}
	return &lastest, nil

}

func FindValidCode(target, scene string) (*model.VerificationCode, error) {
	var v model.VerificationCode
	err := db.Where("target = ?", target).Where("scene = ?", scene).Where("used_at IS NULL").Where("expires_at > ?", time.Now()).Order("created_at DESC").First(&v).Error
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func MarkCodeUsed(id uint) error {
	return db.Model(&model.VerificationCode{}).Where("id = ?", id).Update("used_at", time.Now()).Error

}

func FindByemail(email string) (*model.User, error) {
	var user model.User
	err := db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return &model.User{}, err
	}
	return &user, nil

}

func UpdateEmail(userID uint, email string) error{
	return db.Model(&model.User{}).Where("id=?", userID).Update("email", email).Error


}
