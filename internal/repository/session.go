package repository

import (
	"gin-demo/internal/model"
	"time"
)

func CreateSession(s *model.Session) error {
	return db.Create(s).Error
}

func FindByAccessHash(hash string) (*model.Session, error) {
	var s model.Session
	err := db.Where("access_token_hash = ?", hash).First(&s).Error
	if err != nil {
		return nil, err
	}
	return &s, nil

}

func FindByRefreshHash(hash string) (*model.Session, error) {
	var s model.Session
	err := db.Where("refresh_token_hash = ?", hash).First(&s).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func FindByPrevRefreshHash(hash string) (*model.Session, error) {
	var s model.Session
	err := db.Where("prev_refresh_hash = ?", hash).First(&s).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func RevokeSession(id string) error {
	return db.Model(&model.Session{}).Where("id = ?", id).Update("revoked_at", time.Now()).Error
}

func RevokeAllSessionsForUser(userID uint) error {
	return db.Model(&model.Session{}).Where("user_id = ?", userID).Update("revoked_at", time.Now()).Error
}

func RotateTokens(id, oldRefreshHash, newAccessHash, newRefreshHash string) error {
	return db.Model(&model.Session{}).Where("id = ?", id).Updates(map[string]interface{}{
		"access_token_hash":  newAccessHash,
		"refresh_token_hash": newRefreshHash,
		"prev_refresh_hash":  oldRefreshHash,
	}).Error
}
