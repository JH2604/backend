package service

//Service 干什么？	编排业务：调用 Repository，未来在这里加规则
import "gin-demo/internal/repository"

// GetUnreadCount 业务层：获取某用户的未读私信数（M1）
// userID：当前登录用户的 ID（由 Handler 从 token 解析后传进来）
func GetUnreadCount(userID uint) (int64, error) {
	return repository.CountUnread(userID)
}
