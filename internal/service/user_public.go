package service

import (
	"errors"

	"gin-demo/internal/model"
	"gin-demo/internal/repository"

	"gorm.io/gorm"
)

// ErrUserNotFound U7 要看的人不存在 → handler 翻译成 40400
var ErrUserNotFound = errors.New("用户不存在")

// ListAdmins U6 管理员列表
func ListAdmins(viewerID uint) ([]model.AdminBrief, error) {
	users, err := repository.FindAdmins()
	if err != nil {
		return nil, err
	}
	list := make([]model.AdminBrief, 0, len(users))
	for _, u := range users {
		list = append(list, model.AdminBrief{
			ID:         u.ID,
			Name:       u.Name,
			AvatarURL:  u.AvatarURL,
			Role:       u.Role,
			Email:      u.Email,
			CanMessage: u.ID != viewerID,
		})
	}
	return list, nil
}

// GetUserPublic U7 查看发帖人信息
// viewerRole 是查看者的角色，来自 token（前端伪造不了），由它决定 detail 给不给
func GetUserPublic(targetID uint, viewerID uint, viewerRole string) (*model.UserPublic, error) {
	user, err := repository.FindUserByID(targetID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 数据库里没这行 → 业务错误，handler 才能对应到 404
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
		PostCount:  count,
		CanMessage: targetID != viewerID,
		CanRemind:  canRemind(user),
	}

	// 只有管理员才填 Detail；普通用户保持 nil → JSON 里就是 null
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
