package service

//Service 干什么？	编排业务：调用 Repository，未来在这里加规则
import (
	"errors"

	"gin-demo/internal/model"
	"gin-demo/internal/repository"

	"gorm.io/gorm"
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

// ErrSelfMessage：M3 —— 不能给自己发私信（自己跟自己不存在"会话"这回事）
var ErrSelfMessage = errors.New("不能给自己发私信")

// SendMessage：M3 业务层 —— 发送私信（POST /messages）
// 返回 (落库后的私信, 提醒结果, error)
func SendMessage(senderID uint, req model.SendMessageReq) (*model.Message, model.RemindResult, error) {
	// ① 不能发给自己。文档没有明说，属于常识性守卫 ——
	//    真要放开，删掉这三行即可，不影响其它任何逻辑
	if req.ReceiverID == senderID {
		return nil, model.RemindResult{}, ErrSelfMessage
	}

	// ② 收信人必须真实存在 → handler 翻译成 40400
	receiver, err := repository.FindUserByID(req.ReceiverID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, model.RemindResult{}, ErrUserNotFound
		}
		return nil, model.RemindResult{}, err
	}

	// ③ 先把提醒的结局算出来（含限流判断）—— 它决定 reminded 这一列写不写 true。
	//    文档没要求提醒时，直接给个"没被请求"的结果，不用进 resolveRemind
	remind := model.RemindResult{
		Status: model.RemindStatusSkipped,
		Reason: model.RemindReasonNotRequested,
	}
	if req.Remind {
		// 注意：resolveRemind 不返回 error —— 提醒失败不影响私信
		remind = resolveRemind(senderID, receiver)
	}

	// ④ 落库。私信和提醒是两件独立的事：
	//    remind 无论是 skipped 还是 failed，这一条 INSERT 都照常执行
	m := &model.Message{
		SenderID:   senderID,
		ReceiverID: req.ReceiverID,
		PostID:     req.PostID,
		Content:    req.Content,
		// 只有"真的投递成功"才算提醒过 —— 这是 reminded 列的定义
		Reminded: remind.Status == model.RemindStatusSent,
	}
	if err := repository.CreateMessage(m); err != nil {
		return nil, model.RemindResult{}, err
	}
	return m, remind, nil
}

// ListConversation：M4 业务层 —— 与某用户的私信记录
func ListConversation(userID, peerID uint, q *model.ConversationQuery) (*model.ConversationResult, error) {
	// ① Limit 补默认值 / 封顶（分页边界归 service 管，repository 不补）
	if q.Limit == 0 {
		q.Limit = model.DefaultPageSize
	}
	if q.Limit > model.MaxPageSize {
		q.Limit = model.MaxPageSize
	}

	// ② 对端必须存在 → 40400。
	//    顺便把 can_remind 要用的"提醒开关 + 联系方式"一起拿到手，不用再查第二次
	peer, err := repository.FindUserByID(peerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	// ③ 取一页：repository 按 id 倒序多取 1 条，用那条"探针"判断 has_more
	messages, hasMore, err := repository.ListConversation(userID, peerID, q.BeforeID, q.Limit)
	if err != nil {
		return nil, err
	}

	// ④ 先查后标 —— 这个顺序不能反：
	//    如果先标已读再查列表，返回的 is_read 会全变成 true，
	//    前端就丢了"刚才哪几条是未读"这个信息（红点会消失得莫名其妙）
	if q.MarkRead {
		// 复用 M5 的 repository：只改 sender_id = peerID（他发给我的）、且此刻未读的
		if _, err := repository.MarkRead(userID, nil, &peerID, false); err != nil {
			return nil, err
		}
	}

	// ⑤ 翻成时间正序：repository 取的是 id DESC（游标好写），
	//    聊天气泡要自上而下读，所以这里倒着遍历装一遍
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

// toMessageView：把消息实体翻译成 M3 / M4 共用的视图。
// direction 相对【我】算：我发的是 sent，收到的（peer 发的）是 received
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
