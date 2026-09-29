// 业务：哈希、重名判断
package service

import (
	"errors"
	"gin-demo/internal/model"
	"gin-demo/internal/repository"
	"gin-demo/pkg/config"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var ErrUserExists = errors.New("用户已存在")
var ErrInvalidCredentials = errors.New("用户名或密码错误")
var ErrInvalidToken = errors.New("token 不合法")

func RegisterUser(username, password string) (*model.User, error) {

	haxi, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user := &model.User{
		Username: username,
		Password: string(haxi),
		Role:     model.RoleUser,
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

func CreateToken(user *model.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
		"username": user.Username,
		"role":     user.Role,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.JWTSecret))
}

func ParseToken(tokenStr string) (*model.TokenInfo, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		return []byte(config.JWTSecret), nil
	})
	if err != nil {
		return nil, err
	}
	claims := token.Claims.(jwt.MapClaims)
	rawUserID := claims["user_id"]
	userID := uint(rawUserID.(float64))
	role, ok := claims["role"].(string)
	username, _ := claims["username"].(string)

	if !ok {
		return nil, ErrInvalidToken

	}
	return &model.TokenInfo{
		UserID:   userID,
		Username: username,
		Role:     role,
	}, nil
}
