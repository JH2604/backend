package repository

import "gin-demo/internal/model"

func CreatePost(post *model.Post) error {
	return db.Create(post).Error
}
