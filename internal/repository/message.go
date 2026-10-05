package repository

import (
	"time"

	"gin-demo/internal/model"

	"gorm.io/gorm"
)

func CountUnread(receiverID uint) (int64, error) {

	var count int64
	//`db.Model(&model.Message{})` = "我要操作 `messages` 表"。
	//repository直接对应修改数据库
	//model.message是自己的包
	err := db.Model(&model.Message{}).
		//"条件是……"（`?` 是占位符，防 SQL 注入）
		//读法："从 `messages` 表里，数一数有多少行满足：
		// `receiver_id` 是 5、且 `is_read` 是 false、且没被软删除。"
		Where("receiver_id = ? AND is_read = ?", receiverID, false).
		//如果count不加&那么修改的就是副本，count依旧是0
		Count(&count).Error
	return count, err
}

// MarkRead：M5 把"发给我的"未读消息标记为已读，返回【真正被改动的行数】
// 三种范围互斥，优先级 ids > peerID > all（service 层负责定优先级，这里负责按范围干活）
func MarkRead(receiverID uint, ids []uint, peerID *uint, all bool) (int64, error) {
	// 公共底盘：只能改"发给我的、且此刻未读"的消息
	// receiver_id   → 越权保护，别人的消息我动不了
	// is_read=false → 让 updated 真的表示"从未读变成已读"，
	//                 同时绕开 MySQL 在"值没变化"时 RowsAffected 可能为 0 的坑
	query := db.Model(&model.Message{}).
		Where("receiver_id = ? AND is_read = ?", receiverID, false)

	switch {
	case len(ids) > 0:
		// 点名模式：只改 id 在这个列表里的
		// "IN ?" 后面直接传切片本身，GORM 会自己展开成 (1,3,7)
		query = query.Where("id IN ?", ids)
	case peerID != nil:
		// 会话模式：只改"这个人发给我的"
		// 必须解引用 *peerID，传指针进去 SQL 就对不上了
		query = query.Where("sender_id = ?", *peerID)
	case all:
		// 全部模式：不加额外条件，底盘本身就是"我的 + 未读"
	default:
		// 三个范围都没传：一行都不改，直接返回 0
		// 参数合不合法由 service 判，repository 只负责"有范围才干活"
		// 少了这个 default，空 body 就会等价于 all，一键清空全部未读
		return 0, nil
	}

	// map 是"列名 → 值"的字典，值那栏用 `interface{}` 是因为两列类型不同；
	// 用 map 而不用 struct 是为了躲开零值被跳过的陷阱
	result := query.Updates(map[string]interface{}{
		"is_read": true,
		"read_at": time.Now(),
	})

	// RowsAffected = 真正被改动的行数，service 拿它当 updated 返回给前端
	// 必须先用 result 接住再分字段取，写成 .Updates(...).Error 就把行数丢了
	return result.RowsAffected, result.Error
}

// ListMessages：M2 —— 我的消息列表（GET /messages）
// 返回 (当页消息, 总条数, error)
//
// 约定：q.Page / q.PageSize 由 service 层补默认值，这里不补
// （PageSize 为 0 会拼出 LIMIT 0，列表永远是空的）
// direction / peer / post 这些给前端看的字段由 service 层拼装，这里只返回实体
func ListMessages(q model.MessageListQuery, userID uint) ([]model.Message, int64, error) {
	// ① 哪些消息算我的：三个 box 取值 = 三种 WHERE
	query := db.Model(&model.Message{})
	switch q.Box {
	case model.MessageBoxSent:
		query = query.Where("sender_id = ?", userID)
	case model.MessageBoxReceived:
		query = query.Where("receiver_id = ?", userID)
	default:
		// box 空串 或 all：我发的 + 我收的
		// OR 必须写在同一个字符串里（GORM 才会加括号）。
		// 拆成两个 Where 会变成 sender_id = ? OR (receiver_id = ? AND is_read = ?)，
		// 我发出的消息就绕过了 is_read 筛选
		query = query.Where("sender_id = ? OR receiver_id = ?", userID, userID)
	}

	// ② is_read：nil = 没传这个参数（不筛）；非 nil 才按 true/false 筛
	// 用 *bool 就是为了区分"没传"和"传了 false"
	if q.IsRead != nil {
		query = query.Where("is_read = ?", *q.IsRead)
	}

	// ③ total：符合条件的总条数（不是当页条数）
	// Session(&gorm.Session{}) 克隆一份去 COUNT，避免影响下面的 query
	var total int64
	if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// ④ 当页数据：created_at 倒序（最新在前），再加 id 倒序防同一秒的消息乱序
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
