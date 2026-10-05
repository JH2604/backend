package repository

import (
	"gin-demo/internal/model"
	"time"
)

func CreatePost(post *model.Post) error {
	return db.Create(post).Error
}

func ListPosts(q model.ListPostsQuery, userID uint) ([]model.Post, int64, error) {
	var query = db.Model(&model.Post{})
	if q.Type != "" {
		query = query.Where("type =?", q.Type)
	}
	if q.Status != "" {
		query = query.Where("status =?", q.Status)

	}
	if q.Keyword != "" {
		p := "%" + q.Keyword + "%"
		query = query.Where(
			"title LIKE ? OR content LIKE ? OR JSON_UNQUOTE(JSON_EXTRACT(location, '$.name')) LIKE ?", p, p, p,
		)

	}
	if q.Mine {
		query = query.Where("user_id =?", userID)

	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var posts []model.Post
	a := "created_at"
	err := query.Offset((q.Page - 1) * q.PageSize).Limit(q.PageSize).Order(a + " " + q.Order).Find(&posts).Error
	if err != nil {
		return nil, 0, err
	}
	return posts, total, nil
}

func GetPostByID(id uint) (*model.Post, error) {
	var post model.Post
	err := db.First(&post, id).Error
	if err != nil {
		return nil, err
	}
	return &post, nil
}

// FindPostsByIDs：按 ID 批量取帖子，给消息列表的 post 简介用
func FindPostsByIDs(ids []uint) ([]model.Post, error) {
	// 空列表没必要查（GORM 会拼成 id IN (NULL)，一行也匹配不到）
	if len(ids) == 0 {
		return nil, nil
	}
	var posts []model.Post
	// 只取 id/title（Content 是 TEXT，不用拖出来）；软删除的帖子会被自动过滤
	err := db.Select("id", "title").Where("id IN ?", ids).Find(&posts).Error
	if err != nil {
		return nil, err
	}
	return posts, nil
}

func UpdatePostStatus(id uint, status string, closedAt *time.Time) error {
	return db.Model(&model.Post{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":    status,
		"closed_at": closedAt,
	}).Error
}

func SoftDeletePost(id uint, reason string) error {
	err := db.Model(&model.Post{}).Where("id = ?", id).Update("delete_reason", reason).Error
	if err != nil {
		return err
	}
	err = db.Delete(&model.Post{}, id).Error
	if err != nil {
		return err
	}
	return nil
}

func CountPost(userID uint) (number int64, err error) {
	err = db.Model(&model.Post{}).Where("user_id=?", userID).Count(&number).Error
	if err != nil {
		return 0, err
	}
	return number, nil
}
