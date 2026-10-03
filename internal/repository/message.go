package repository

import "gin-demo/internal/model"

func CountUnread(receiverID uint) (int64, error) {

	var count int64
	//`db.Model(&model.Message{})` = "我要操作 `messages` 表"。
	err := db.Model(&model.Message{}).
		Where("receiver_id = ? AND is_read = ?", receiverID, false).
		Count(&count).Error
	return count, err
}
