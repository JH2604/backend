// 业务：哈希、重名判断
package service

import (
	"errors"
	"gin-demo/internal/model"
	"gin-demo/internal/repository"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var ErrUserExists = errors.New("用户已存在")
var ErrInvalidCredentials = errors.New("用户名或密码错误")
var ErrInvalidToken = errors.New("token 不合法")
var ErrSessionInvalid = errors.New("会话已失效")
var ErrRefreshTokenInvalid = errors.New("刷新令牌无效")
var ErrTokenExpired = errors.New("令牌已过期")

func RegisterUser(username, password string) (*model.User, error) {

	haxi, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user := &model.User{
		Username: username,
		Password: string(haxi),
		Role:     model.RoleStudent,
	}
	err = repository.CreateUser(user)
	if err != nil {
		reception := strings.Contains(err.Error(), "Duplicate entry")
		if reception {
			return nil, ErrUserExists
		}
		return nil, err

	}
	return user, nil
}

func LoginUser(username, password string) (*model.User, error) {
	user, err := repository.FindByUsername(username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, ErrInvalidCredentials
	}
	return &user, nil

}

// 修改密码函数
func ChangePassword(userID uint, old_password string, newpassword string) error {
	user, err := repository.FindUserByID(userID)
	if err != nil {
		return err
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(old_password))
	if err != nil {
		return ErrOldPasswordWrong

	}
	new_password, err := bcrypt.GenerateFromPassword([]byte(newpassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	err = repository.SavePassword(user.ID, string(new_password))
	if err != nil {
		return err
	}
	err = repository.RevokeAllSessionsForUser(user.ID)
	if err != nil {
		return err
	}
	return nil
}
