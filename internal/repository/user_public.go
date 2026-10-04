package repository

import (
	"errors"
	"time"

	"gin-demo/internal/model"

	"gorm.io/gorm"
)

// FindAdmins：U6 —— 取出全部管理员
// Order("id ASC") 别省：MySQL 不保证返回顺序，不加排序前端列表可能「跳来跳去」
func FindAdmins() ([]model.User, error) {
	var users []model.User
	err := db.Where("role = ?", model.RoleAdmin).Order("id ASC").Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

// LastLoginAt：U7 —— 该用户最近一次登录时间
// 依据：每次登录成功都会写一条会话（service.GenerateTokenPair → repository.CreateSession），
// 所以「最新一条会话的 created_at」≈ 最近一次登录时间。
// 表名是 sessions（model.Session 没写 TableName()，GORM 默认按复数命名），不是 user_sessions。
// 一条会话都没有（比如查一个从没登录过的账号）→ 返回 (nil, nil)，
// 前端 JSON 里就是 null，这不算错误。
func LastLoginAt(userID uint) (*time.Time, error) {
	//`users` 表里根本没有 `last_login_at` 这一列，但可通过sessions表获得最新登录时间
	var s model.Session
	//`Order("created_at DESC").First(&s)` = 取最新那一条
	err := db.Where("user_id = ?", userID).Order("created_at DESC").First(&s).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			//用户存在，但从没登录过
			return nil, nil
		}
		//数据库连不上
		return nil, err
	}
	//返回的第一个值需要是指针类型
	return &s.CreatedAt, nil
}
