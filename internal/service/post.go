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

func ListPosts(userID uint, q *model.ListPostsQuery) ([]model.Post, int64, error) {
	if q.Page == 0 {
		q.Page = model.DefaultPage
	}
	if q.PageSize == 0 {
		q.PageSize = model.DefaultPageSize
	}
	if q.Order == "" {
		q.Order = model.DefaultOrder
	}
	if q.PageSize > model.MaxPageSize {
		q.PageSize = model.MaxPageSize
	}
	posts,total,err := repository.ListPosts(*q, userID)
	if err != nil{
		return nil,0,err
	}

	err = attachAuthors(posts)
	if err != nil{
		return nil,0,err
	}

	return	posts,total,nil

}

func attachAuthors(posts []model.Post) error {
	seen := map[uint]bool{}
	ids := []uint{}
	for _, v := range posts {
		if !seen[v.UserID] {
			seen[v.UserID] = true
			ids = append(ids, v.UserID)
		}
	}
	users, err := repository.FindUsersByIDs(ids)
	if err != nil {
		return err
	}
	authorMap := map[uint]model.UserBrief{}
	for _, u := range users {
		authorMap[u.ID] = model.UserBrief{
			Name: u.Username,
			ID: u.ID,
			Role: u.Role,
			AvatarURL: "",

		}
	}
	for p,_ := range posts {
		b,ok := authorMap[posts[p].UserID]
		if ok{
			posts[p].Author = &b
		}
	}
	return nil
}
