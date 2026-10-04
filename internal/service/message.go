package service

//Service 干什么？	编排业务：调用 Repository，未来在这里加规则
import (
	"errors"
	"gin-demo/internal/model"
	"gin-demo/internal/repository"
)

// ErrNoMarkReadScope：M5 三个范围一个都没传（ids 空、peer_id 空、all=false）
// 这一刀必须由 service 拦住 —— 拦不住的话 repository 那边会等价于 all，直接清空全部未读
// errors 包里的 New 函数，给我造一个新的错误，描述是这串字。
var ErrNoMarkReadScope = errors.New("未指定标记已读的范围")

// GetUnreadCount 业务层：获取某用户的未读私信数（M1）
// userID：当前登录用户的 ID（由 Handler 从 token 解析后传进来）
func GetUnreadCount(userID uint) (int64, error) {
	return repository.CountUnread(userID)
}

// MarkRead 业务层：标记已读（M5）
// 返回值顺序：(真正改动的条数 updated, 标记后剩余的未读数 unread_total)
// 优先级 ids > peer_id > all，三个都没传就报参数错误，不往下走
func MarkRead(userID uint, req model.MarkReadReq) (int64, int64, error) {
	// ① 三空守卫：必须在调 repository 【之前】拦，
	//    否则那个请求等价于"清空我的全部未读"
	if len(req.IDs) == 0 && req.PeerID == nil && !req.All {
		return 0, 0, ErrNoMarkReadScope
	}

	// ② 先改数据：updated = 真正从"未读"变成"已读"的条数
	//    优先级链不用在这里写 if/else —— repository 的 switch 顺序就是 ids > peerID > all，
	//    三个字段原样透传下去即可
	updated, err := repository.MarkRead(userID, req.IDs, req.PeerID, req.All)
	if err != nil {
		return 0, 0, err
	}

	// ③ 改完【再】数未读：这时拿到的才是"标记之后"的剩余未读数，前端红点才能正确消失
	//    顺序反了就错：先 COUNT 会把刚刚被标记掉的那几条也算进去，返回一个过期的未读数
	unreadTotal, err := repository.CountUnread(userID)
	if err != nil {
		return 0, 0, err
	}
	return updated, unreadTotal, nil
}
