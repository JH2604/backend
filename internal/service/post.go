package service

import (
	"errors"
	"gin-demo/internal/model"
	"gin-demo/internal/repository"
	"time"

	"gorm.io/gorm"
)

var ErrPostNotFound = errors.New("帖子不存在")
var ErrNotPostOwner = errors.New("不是本人的帖子")
var ErrNoPermission = errors.New("没有权限")


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
	p.IsMine = true
	p.CanDelete = true
	p.CanChangeStatus = true

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
	posts, total, err := repository.ListPosts(*q, userID)
	if err != nil {
		return nil, 0, err
	}

	err = attachAuthors(posts)
	if err != nil {
		return nil, 0, err
	}

	return posts, total, nil

}

// toUserBrief 把 User 转成对外的 UserBrief。
// ⚠️ 只挑这 4 个字段 —— 绝不能带上 Password，那是 bcrypt 哈希
func toUserBrief(u model.User) model.UserBrief {
	return model.UserBrief{
		ID:        u.ID,
		Name:      u.Username,
		AvatarURL: "",
		Role:      u.Role,
	}
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
		authorMap[u.ID] = toUserBrief(u)
	}
	for p, _ := range posts {
		b, ok := authorMap[posts[p].UserID]
		if ok {
			posts[p].Author = &b
		}
	}
	return nil
}

func UpdatePostStatus(postID, userID uint, status string) (*model.Post, error) {
	post, err := repository.GetPostByID(postID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPostNotFound
		}
		return nil, err
	}
	if post.UserID != userID {
		return nil, ErrNotPostOwner
	}
	if post.Status == status {
		return post, nil
	}
	var closedAt *time.Time // 默认 nil —— 撤回时正好要写 NULL
	if status == model.PostStatusClosed {
		now := time.Now()
		closedAt = &now
	}
	err = repository.UpdatePostStatus(postID, status, closedAt)
	if err != nil {
		return nil, err
	}
	post.Status = status
	post.ClosedAt = closedAt
	return post, nil

}

func GetPost(postID, userID uint, role string) (*model.Post, error) {
	post, err := repository.GetPostByID(postID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPostNotFound
		}
		return nil, err
	}
	v, err := repository.FindUserByID(post.UserID)
	if err != nil {
		return nil, err
	}
	brief := toUserBrief(*v)
	post.Author = &brief
	post.IsMine = post.UserID == userID
	post.CanDelete = post.IsMine || role == model.RoleAdmin
	post.CanChangeStatus = post.IsMine
	return post,nil

}

func DeletePost(postID,userID uint,role,reason string)error{
	post,err := repository.GetPostByID(postID)
	if err != nil {
    	if errors.Is(err, gorm.ErrRecordNotFound) {
        	return ErrPostNotFound
    }
    return err
}
	if post.UserID != userID && role != model.RoleAdmin{
		return ErrNoPermission
	}
	err = repository.SoftDeletePost(postID,reason)
	if err != nil{
		return err
	}
	return nil
}