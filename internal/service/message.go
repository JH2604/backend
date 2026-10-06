package service

import (
	"errors"

	"gin-demo/internal/model"
	"gin-demo/internal/repository"

	"gorm.io/gorm"
)

// ErrNoMarkReadScope M5 三个范围都没传
var ErrNoMarkReadScope = errors.New("未指定标记已读的范围")

// GetUnreadCount M1 未读私信数
func GetUnreadCount(userID uint) (int64, error) {
	return repository.CountUnread(userID)
}

// MarkRead M5 标记已读，返回 (改动条数, 剩余未读数)；优先级 ids > peer_id > all
func MarkRead(userID uint, req model.MarkReadReq) (int64, int64, error) {
	// 必须先拦：否则空请求等价于"清空我的全部未读"
	if len(req.IDs) == 0 && req.PeerID == nil && !req.All {
		return 0, 0, ErrNoMarkReadScope
	}

	updated, err := repository.MarkRead(userID, req.IDs, req.PeerID, req.All)
	if err != nil {
		return 0, 0, err
	}

	// 改完再数未读：顺序反了会把刚标记掉的几条也算进去，返回过期的未读数
	unreadTotal, err := repository.CountUnread(userID)
	if err != nil {
		return 0, 0, err
	}
	return updated, unreadTotal, nil
}

// ListMessages M2 我的消息列表
func ListMessages(userID uint, q *model.MessageListQuery) ([]model.MessageListItem, int64, error) {
	if q.Page == 0 {
		q.Page = model.DefaultPage
	}
	if q.PageSize == 0 {
		q.PageSize = model.DefaultPageSize
	}
	if q.PageSize > model.MaxPageSize {
		q.PageSize = model.MaxPageSize
	}

	messages, total, err := repository.ListMessages(*q, userID)
	if err != nil {
		return nil, 0, err
	}

	// 批量补 peer / post：各一次 IN 查询，顶掉 N+1
	peerMap, err := loadPeers(messages, userID)
	if err != nil {
		return nil, 0, err
	}
	postMap, err := loadPosts(messages)
	if err != nil {
		return nil, 0, err
	}

	// 用 make 而不是 var：空列表要序列化成 []，不是 null
	items := make([]model.MessageListItem, 0, len(messages))
	for _, m := range messages {
		items = append(items, toMessageItem(m, userID, peerMap, postMap))
	}
	return items, total, nil
}

// directionOf 这一行相对"我"是发出还是收到
func directionOf(m model.Message, userID uint) string {
	if m.SenderID == userID {
		return model.MessageDirectionSent
	}
	return model.MessageDirectionReceived
}

// peerIDOf 这一行的对端是谁：我发的看收件人，我收的看发件人
func peerIDOf(m model.Message, userID uint) uint {
	if directionOf(m, userID) == model.MessageDirectionSent {
		return m.ReceiverID
	}
	return m.SenderID
}

// loadPeers 一次查出这一页的对端用户，做成 id → UserBrief 字典
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

// loadPosts 一次查出这一页的关联帖子；帖子被软删则查不到，该项 post 输出 null
func loadPosts(messages []model.Message) (map[uint]model.MessagePostBrief, error) {
	ids := []uint{}
	seen := map[uint]bool{}
	for _, m := range messages {
		if m.PostID == nil { // nil = 不是从帖子详情页发起的
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

// toMessageItem 消息实体 → M2 列表项
func toMessageItem(m model.Message, userID uint, peers map[uint]model.UserBrief, posts map[uint]model.MessagePostBrief) model.MessageListItem {
	item := model.MessageListItem{
		ID:        m.ID,
		Direction: directionOf(m, userID),
		Content:   m.Content,
		IsRead:    m.IsRead,
		Reminded:  m.Reminded,
		CreatedAt: m.CreatedAt,
	}
	// 查到了才填，查不到留 nil → JSON 里是 null
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

// ErrSelfMessage 不能给自己发私信
var ErrSelfMessage = errors.New("不能给自己发私信")

// SendMessage M3 发送私信，返回 (落库结果, 提醒结果, error)
func SendMessage(senderID uint, req model.SendMessageReq) (*model.Message, model.RemindResult, error) {
	if req.ReceiverID == senderID {
		return nil, model.RemindResult{}, ErrSelfMessage
	}

	receiver, err := repository.FindUserByID(req.ReceiverID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, model.RemindResult{}, ErrUserNotFound
		}
		return nil, model.RemindResult{}, err
	}

	// 未请求提醒时不进 resolveRemind，直接 skipped / not_requested
	remind := model.RemindResult{
		Status: model.RemindStatusSkipped,
		Reason: model.RemindReasonNotRequested,
	}
	if req.Remind {
		remind = resolveRemind(senderID, receiver)
	}

	// 私信与提醒相互独立：remind 为 skipped/failed 时 INSERT 照常执行
	m := &model.Message{
		SenderID:   senderID,
		ReceiverID: req.ReceiverID,
		PostID:     req.PostID,
		Content:    req.Content,
		Reminded:   remind.Status == model.RemindStatusSent, // 只有真正投递成功才算提醒过
	}
	if err := repository.CreateMessage(m); err != nil {
		return nil, model.RemindResult{}, err
	}
	return m, remind, nil
}

// ListConversation M4 与某用户的私信记录（游标分页）
func ListConversation(userID, peerID uint, q *model.ConversationQuery) (*model.ConversationResult, error) {
	if q.Limit == 0 {
		q.Limit = model.DefaultPageSize
	}
	if q.Limit > model.MaxPageSize {
		q.Limit = model.MaxPageSize
	}

	peer, err := repository.FindUserByID(peerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	messages, hasMore, err := repository.ListConversation(userID, peerID, q.BeforeID, q.Limit)
	if err != nil {
		return nil, err
	}

	// 先查后标：先标已读会让返回的 is_read 全变 true，丢失"哪几条原本未读"
	if q.MarkRead {
		if _, err := repository.MarkRead(userID, nil, &peerID, false); err != nil {
			return nil, err
		}
	}

	// repository 返回 id 倒序，这里翻成时间正序
	items := make([]model.MessageView, 0, len(messages))
	for i := len(messages) - 1; i >= 0; i-- {
		items = append(items, toMessageView(messages[i], userID))
	}

	return &model.ConversationResult{
		Peer:      toUserBrief(*peer),
		CanRemind: canRemind(peer),
		List:      items,
		HasMore:   hasMore,
	}, nil
}

// toMessageView 消息实体 → M3/M4 共用视图
func toMessageView(m model.Message, userID uint) model.MessageView {
	return model.MessageView{
		ID:        m.ID,
		Direction: directionOf(m, userID),
		Content:   m.Content,
		IsRead:    m.IsRead,
		Reminded:  m.Reminded,
		CreatedAt: m.CreatedAt,
	}
}
