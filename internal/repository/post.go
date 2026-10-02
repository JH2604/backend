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
		query = query.Where("title LIKE ? OR content LIKE ?", "%"+q.Keyword+"%", "%"+q.Keyword+"%")

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
	if model.PostSortFields[q.SortBy] {
		a = q.SortBy
	}
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
func UpdatePostStatus(id uint, status string, closedAt *time.Time) error {
	return db.Model(&model.Post{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":    status,
		"closed_at": closedAt,
	}).Error
}
