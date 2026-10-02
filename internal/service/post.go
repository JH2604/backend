package service

import (
	"gin-demo/internal/model"
	"gin-demo/internal/repository"
)

func CreatePost(userID uint, req model.CreatePostReq) (*model.Post, error) {
	p := &model.Post{
		Type:      req.Type,
		UserID:    userID,
		Title:     req.Title,
		Content:   req.Content,
		Images:    req.Images,
		Location:  req.Location,
		EventTime: req.EventTime,
		Status:    model.PostStatusOpen,
	}
	err := repository.CreatePost(p)
	if err != nil {
		return nil, err
	}
	return p, nil
}
