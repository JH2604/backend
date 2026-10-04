package service

import (
	"errors"

	"gin-demo/internal/model"
	"gin-demo/internal/repository"

	"gorm.io/gorm"
)

// ErrUserNotFound：U7 要看的人不存在 → handler 翻译成 40400
var ErrUserNotFound = errors.New("用户不存在")

// ListAdmins：U6「联系管理员」
func ListAdmins() ([]model.AdminBrief, error) {
	users, err := repository.FindAdmins()
	if err != nil {
		return nil, err
	}
	// 这里用 make(..., 0, len(users)) 而不是 var list []model.AdminBrief：
	// 空切片序列化成 []，nil 切片序列化成 null —— 前端要的是 []
	list := make([]model.AdminBrief, 0, len(users))
	for _, u := range users {
		list = append(list, model.AdminBrief{
			ID:        u.ID,
			Name:      u.Name,
			AvatarURL: u.AvatarURL,
			Role:      u.Role,
			Email:     u.Email,
			// 系统目前没有「拉黑 / 禁止私信」这个概念，所以恒为 true；
			// 以后真要做黑名单，只改这一行
			CanMessage: true,
		})
	}
	return list, nil
}

// GetUserPublic：U7 查看发帖人信息
// viewerRole 是【查看者】的角色，来自 token（前端伪造不了），由它决定 detail 给不给
func GetUserPublic(targetID uint, viewerRole string) (*model.UserPublic, error) {
	user, err := repository.FindUserByID(targetID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 把「数据库里没这行」翻译成业务错误，handler 才能对应到 404
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	count, err := repository.CountPost(targetID)
	if err != nil {
		return nil, err
	}

	out := &model.UserPublic{
		ID:        user.ID,
		Name:      user.Name,
		AvatarURL: user.AvatarURL,
		Role:      user.Role,
		PostCount: count,
		// 同 U6：暂无黑名单，恒 true
		CanMessage: true,
		// 对方自己关了提醒开关、或者手机邮箱都没绑 → 提醒不了
		CanRemind: user.AllowRemind && (user.Phone != "" || user.Email != ""),
	}

	// ★ 关键的一刀：只有管理员才填 Detail。
	// 普通用户这里不赋值，Detail 保持 nil → JSON 里就是 null
	if viewerRole == model.RoleAdmin {
		lastLogin, err := repository.LastLoginAt(targetID)
		if err != nil {
			return nil, err
		}
		out.Detail = &model.UserDetail{
			StudentID:   user.StudentID,
			Phone:       user.Phone,
			Email:       user.Email,
			AllowRemind: user.AllowRemind,
			CreatedAt:   user.CreatedAt,
			LastLoginAt: lastLogin,
		}
	}
	return out, nil
}
