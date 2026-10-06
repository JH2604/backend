// 业务：哈希、重名判断
package service

import (
	"errors"
	"gin-demo/internal/model"
	"gin-demo/internal/repository"
	"gin-demo/pkg/validate"
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
var ErrStudentNotFound = errors.New("没找到学生")
var ErrInvalidParams = errors.New("参数错误")
var ErrCodeInvalid = errors.New("验证码错误或过期")
var ErrContactTaken = errors.New("邮箱已被其他账号绑定")


func RegisterUser(studentID, password, role string) (*model.User, error) {
	studentID = strings.TrimSpace(studentID)
	role = strings.TrimSpace(role)
	if !validate.StudentID(studentID) || !validate.Password(password) {
		return nil, ErrInvalidParams
	}
	if role != model.RoleStudent && role != model.RoleAdmin {
		return nil, ErrInvalidParams
	}
	student, err := repository.FindStudent(studentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrStudentNotFound
		}
		return nil, err
	}
	haxi, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user := &model.User{
		StudentID:   student.StudentID,
		Password:    string(haxi),
		Role:        role,
		Name:        student.StudentName,
		Theme:       "light",
		AllowRemind: true,
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
	user, err := repository.FindByStudentID(username)
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
	if !validate.Password(newpassword) {
		return ErrInvalidParams
	}
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

func GetUserProfile(userID uint) (*model.User, error) {
	user, err := repository.FindUserByID(userID)
	if err != nil {
		return nil, err
	}
	count, err := repository.CountPost(userID)
	if err != nil {
		return nil, err
	}
	user.PostCount = count
	return user, nil
}

func UpdateAvatar(userID uint, avatarURL string) (*model.User, error) {
	avatarURL = strings.TrimSpace(avatarURL)
	if !strings.HasPrefix(avatarURL, "/uploads/") {
		return nil, ErrInvalidParams
	}
	if err := repository.UpdateAvatar(userID, avatarURL); err != nil {
		return nil, err
	}
	return GetUserProfile(userID)
}
