package repository

import "gin-demo/internal/model"

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
