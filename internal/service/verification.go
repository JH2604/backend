package service

import (
	"encoding/binary"
	"errors"
	"fmt"
	"gin-demo/internal/mail"
	"gin-demo/internal/model"
	"gin-demo/internal/repository"
	"time"

	"gorm.io/gorm"
)

var ErrCodeTooFrequent = errors.New("请求太频繁")

func SendVerificationCode(target, scene string) error {
	latest, err := repository.FindLatestByTargetScene(target, scene)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {

		} else {
			return err
		}
	}
	if latest != nil {
		if time.Since(latest.CreatedAt) < 60*time.Second {
			return ErrCodeTooFrequent
		}
	}
	buf, err := randomBytes(4)
	if err != nil {
		return err
	}
	n := binary.BigEndian.Uint32(buf)
	n = n % 1000000
	code := fmt.Sprintf("%06d", n)
	v := &model.VerificationCode{
		Target:    target,
		Scene:     scene,
		Code:      code,
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}
	err = repository.CreateVerificationCode(v)
	if err != nil {
		return err
	}
	err = mail.Default.SendVerificationCode(target, code)
	if err != nil {
		return err
	}
	return nil
}

func BindContact(userID uint, target,scene, code string) error {
	target1, err := repository.FindValidCode(target, scene)
	if err != nil{
		return	ErrCodeInvalid
	}
	if target1.Code != code{
		return ErrCodeInvalid
	}
	u,err := repository.FindByemail(target)
	if err == nil && u.ID != userID {
    	return ErrContactTaken        // 别人绑了
	}
	err = repository.UpdateEmail(userID,target)
	if err != nil{
		return err
	}
	err = repository.MarkCodeUsed(target1.ID)
	if err != nil{
		return err
	}
	return nil
}
