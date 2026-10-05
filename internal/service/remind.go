package service

import (
	"fmt"
	"time"

	"gin-demo/internal/model"
	"gin-demo/internal/repository"
)

// M3 提醒的两条限流红线（文档原文）
const (
	remindPeerCooldown = 12 * time.Hour // 同一对用户 12 小时内只提醒 1 次
	remindDailyLimit   = 5              // 每个接收人每天最多被提醒 5 次
)

// canRemind 能否提醒：对方开了开关且至少绑了手机或邮箱（U7 的 can_remind 与 M4 共用）
func canRemind(u *model.User) bool {
	return u.AllowRemind && (u.Phone != "" || u.Email != "")
}

// resolveRemind 提醒的全部决策，按"该不该发 → 发得出去吗 → 投递"排序。
// 故意不返回 error：提醒失败不影响私信本身，异常就地转成 failed/provider_error
func resolveRemind(senderID uint, receiver *model.User) model.RemindResult {
	skip := func(reason string) model.RemindResult {
		return model.RemindResult{Status: model.RemindStatusSkipped, Reason: reason}
	}
	fail := func(channel string) model.RemindResult {
		return model.RemindResult{Status: model.RemindStatusFailed, Channel: channel, Reason: model.RemindReasonProviderErr}
	}

	if !receiver.AllowRemind {
		return skip(model.RemindReasonDisabled)
	}
	if receiver.Phone == "" && receiver.Email == "" {
		return skip(model.RemindReasonNoContact)
	}

	// 限流一：同对 12 小时
	since := time.Now().Add(-remindPeerCooldown)
	samePair, err := repository.CountRemindedTo(senderID, receiver.ID, since)
	if err != nil {
		// 查不出限流状态时宁可漏发也不多发
		fmt.Println("❌ 提醒限流查询失败（对端维度）:", err)
		return fail("")
	}
	if samePair > 0 {
		return skip(model.RemindReasonRateLimited)
	}

	// 限流二：接收人每日上限
	sentToday, err := repository.CountRemindedForReceiver(receiver.ID, startOfToday())
	if err != nil {
		fmt.Println("❌ 提醒限流查询失败（日维度）:", err)
		return fail("")
	}
	if sentToday >= remindDailyLimit {
		return skip(model.RemindReasonRateLimited)
	}

	channel, err := sendRemindNotice(receiver)
	if err != nil {
		fmt.Println("❌ 提醒投递失败:", err)
		return fail(channel)
	}
	return model.RemindResult{Status: model.RemindStatusSent, Channel: channel}
}

// startOfToday 今天 0 点（本地时区），"每天 5 次"按自然日算
func startOfToday() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
}

// sendRemindNotice 投递提醒 —— 目前是桩：项目还没有短信/邮件通道，只打脱敏日志。
// 接入真实服务商时只改这里，返回 error 即变成 failed/provider_error
func sendRemindNotice(receiver *model.User) (string, error) {
	if receiver.Phone != "" { // 手机优先，其次邮箱
		fmt.Printf("📨 [提醒桩] 手机提醒 → 用户 #%d，号码 %s\n", receiver.ID, maskContact(receiver.Phone))
		return model.RemindChannelSMS, nil
	}
	fmt.Printf("📨 [提醒桩] 邮箱提醒 → 用户 #%d，地址 %s\n", receiver.ID, maskContact(receiver.Email))
	return model.RemindChannelEmail, nil
}

// maskContact 手机号/邮箱脱敏，日志不留明文
func maskContact(s string) string {
	r := []rune(s)
	switch {
	case len(r) <= 2:
		return "**"
	case len(r) <= 8:
		return string(r[:1]) + "****" + string(r[len(r)-1:])
	default:
		return string(r[:3]) + "****" + string(r[len(r)-3:])
	}
}
