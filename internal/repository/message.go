package repository

import (
	"time"

	"gin-demo/internal/model"

	"gorm.io/gorm"
)

func CountUnread(receiverID uint) (int64, error) {
	var count int64
	err := db.Model(&model.Message{}).
		Where("receiver_id = ? AND is_read = ?", receiverID, false).
		Count(&count).Error
	return count, err
}

// MarkRead M5 标记已读，返回真正被改动的行数；范围优先级 ids > peerID > all
func MarkRead(receiverID uint, ids []uint, peerID *uint, all bool) (int64, error) {
	// 底盘：只能改"发给我的、且此刻未读"的
	query := db.Model(&model.Message{}).
		Where("receiver_id = ? AND is_read = ?", receiverID, false)

	switch {
	case len(ids) > 0:
		query = query.Where("id IN ?", ids)
	case peerID != nil:
		query = query.Where("sender_id = ?", *peerID)
	case all:
	default:
		// 三个范围都没传：一行都不改（少了这个 default，空 body 就等于 all）
		return 0, nil
	}

	result := query.Updates(map[string]interface{}{
		"is_read": true,
		"read_at": time.Now(),
	})
	return result.RowsAffected, result.Error
}

// ListMessages M2 我的消息列表：返回 (当页消息, 总条数, error)
// 分页默认值由 service 补，peer/post 由 service 拼装
func ListMessages(q model.MessageListQuery, userID uint) ([]model.Message, int64, error) {
	query := db.Model(&model.Message{})
	switch q.Box {
	case model.MessageBoxSent:
		query = query.Where("sender_id = ?", userID)
	case model.MessageBoxReceived:
		query = query.Where("receiver_id = ?", userID)
	default:
		// OR 必须写在同一个字符串里 GORM 才会加括号，拆成两个 Where 会绕过 is_read 筛选
		query = query.Where("sender_id = ? OR receiver_id = ?", userID, userID)
	}

	if q.IsRead != nil {
		query = query.Where("is_read = ?", *q.IsRead)
	}

	var total int64
	if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var messages []model.Message
	err := query.
		Order("created_at DESC, id DESC").
		Offset((q.Page - 1) * q.PageSize).
		Limit(q.PageSize).
		Find(&messages).Error
	if err != nil {
		return nil, 0, err
	}

	return messages, total, nil
}

// CreateMessage 落库一条私信
func CreateMessage(m *model.Message) error {
	return db.Create(m).Error
}

// ListConversation 会话内私信，游标分页；返回 (倒序结果, 是否还有更早, error)
func ListConversation(userID, peerID uint, beforeID *uint, limit int) ([]model.Message, bool, error) {
	// 双向条件必须整体加括号，否则游标只作用于后半个 OR，翻页会重复
	query := db.Model(&model.Message{}).
		Where("(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)",
			userID, peerID, peerID, userID)

	if beforeID != nil {
		query = query.Where("id < ?", *beforeID)
	}

	var messages []model.Message
	if err := query.Order("id DESC").Limit(limit + 1).Find(&messages).Error; err != nil {
		return nil, false, err
	}

	hasMore := len(messages) > limit
	if hasMore {
		messages = messages[:limit] // 多取的那 1 条不返回
	}
	return messages, hasMore, nil // 倒序，翻正序在 service 做
}

// CountRemindedTo 我成功提醒过该用户的次数（限流用）
func CountRemindedTo(senderID, receiverID uint, since time.Time) (int64, error) {
	var count int64
	err := db.Model(&model.Message{}).
		Where("sender_id = ? AND receiver_id = ? AND reminded = ? AND created_at >= ?",
			senderID, receiverID, true, since).
		Count(&count).Error
	return count, err
}

// CountRemindedForReceiver 该用户被成功提醒的总次数（限流用）
func CountRemindedForReceiver(receiverID uint, since time.Time) (int64, error) {
	var count int64
	err := db.Model(&model.Message{}).
		Where("receiver_id = ? AND reminded = ? AND created_at >= ?", receiverID, true, since).
		Count(&count).Error
	return count, err
}
