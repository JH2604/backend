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

// ListMessages：M2 业务层 —— 我的消息列表（GET /messages）
// 返回 (这一页的列表项, 总条数, error)
// 思路：先查出实体，再【批量】补 peer / post，最后拼成前端要的字段（不在循环里一条条查库）
func ListMessages(userID uint, q *model.MessageListQuery) ([]model.MessageListItem, int64, error) {
	// ① 补分页默认值（分页边界由 service 负责，repository 不管）
	if q.Page == 0 {
		q.Page = model.DefaultPage
	}
	if q.PageSize == 0 {
		q.PageSize = model.DefaultPageSize
	}
	if q.PageSize > model.MaxPageSize {
		q.PageSize = model.MaxPageSize
	}
	// box 不用转换：不传 / all 在 repository 里都是"我发的 + 我收的"

	// ② 查这一页的消息实体（原始数据，还没加工）
	messages, total, err := repository.ListMessages(*q, userID)
	if err != nil {
		return nil, 0, err
	}

	// ③ 批量补对端用户、关联帖子：各一次 IN 查询，顶掉 N+1
	peerMap, err := loadPeers(messages, userID)
	if err != nil {
		return nil, 0, err
	}
	postMap, err := loadPosts(messages)
	if err != nil {
		return nil, 0, err
	}

	// ④ 拼 DTO。用 make 而不是 var：空列表要序列化成 []，不是 null
	items := make([]model.MessageListItem, 0, len(messages))
	for _, m := range messages {
		items = append(items, toMessageItem(m, userID, peerMap, postMap))
	}
	return items, total, nil
}

// directionOf：这一行相对「我」是发出去的，还是收到的
func directionOf(m model.Message, userID uint) string {
	if m.SenderID == userID {
		return model.MessageDirectionSent
	}
	return model.MessageDirectionReceived
}

// peerIDOf：这一行的「对端」是谁 —— 我发的看收件人，我收的看发件人
func peerIDOf(m model.Message, userID uint) uint {
	if directionOf(m, userID) == model.MessageDirectionSent {
		return m.ReceiverID
	}
	return m.SenderID
}

// loadPeers：把这一页涉及的对端用户一次查出来，做成 id → UserBrief 的字典
func loadPeers(messages []model.Message, userID uint) (map[uint]model.UserBrief, error) {
	ids := []uint{}
	seen := map[uint]bool{}
	for _, m := range messages {
		peerID := peerIDOf(m, userID)
		if !seen[peerID] {
			seen[peerID] = true
			ids = append(ids, peerID)
		}
	}
	// 空 ids 时 repository 直接返回，不会查库
	users, err := repository.FindUsersByIDs(ids)
	if err != nil {
		return nil, err
	}
	peerMap := map[uint]model.UserBrief{}
	for _, u := range users {
		peerMap[u.ID] = toUserBrief(u)
	}
	return peerMap, nil
}

// loadPosts：把这一页涉及的关联帖子一次查出来，做成 id → {id, title} 的字典
// 帖子被软删除 → 查不到 → 字典里没有 → 这一项 post 输出 null（消息本身照常返回）
func loadPosts(messages []model.Message) (map[uint]model.MessagePostBrief, error) {
	ids := []uint{}
	seen := map[uint]bool{}
	for _, m := range messages {
		// PostID 是 *uint：nil = 不是从帖子详情页发起的，没有关联帖子
		if m.PostID == nil {
			continue
		}
		if !seen[*m.PostID] {
			seen[*m.PostID] = true
			ids = append(ids, *m.PostID)
		}
	}
	posts, err := repository.FindPostsByIDs(ids)
	if err != nil {
		return nil, err
	}
	postMap := map[uint]model.MessagePostBrief{}
	for _, p := range posts {
		postMap[p.ID] = model.MessagePostBrief{ID: p.ID, Title: p.Title}
	}
	return postMap, nil
}

// toMessageItem：把一条消息实体翻译成前端要的列表项
// is_read 原样透传：我收到的 = 我读没读；我发出的 = 对方读没读（前端自己按 direction 理解）
func toMessageItem(m model.Message, userID uint, peers map[uint]model.UserBrief, posts map[uint]model.MessagePostBrief) model.MessageListItem {
	item := model.MessageListItem{
		ID:        m.ID,
		Direction: directionOf(m, userID),
		Content:   m.Content,
		IsRead:    m.IsRead,
		Reminded:  m.Reminded,
		CreatedAt: m.CreatedAt,
	}
	// 查到了才填；查不到就留 nil → JSON 里是 null
	if peer, ok := peers[peerIDOf(m, userID)]; ok {
		item.Peer = &peer
	}
	if m.PostID != nil {
		if post, ok := posts[*m.PostID]; ok {
			item.Post = &post
		}
	}
	return item
}
