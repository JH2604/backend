package service

import (
	"errors"
	"gin-demo/internal/model"
	"gin-demo/internal/repository"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

var ErrUserExists = errors.New("用户已存在")

func RegisterUser(username, password string) (*model.User, error) {

	haxi, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user := &model.User{
		Username: username,
		Password: string(haxi),
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
