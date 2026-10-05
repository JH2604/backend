package service

import (
	"fmt"
	"time"

	"gin-demo/internal/model"
	"gin-demo/internal/repository"
)

// M3 提醒的两条限流红线（文档原文）
const (
	// 同一发送人 → 同一接收人，12 小时内只提醒 1 次
	remindPeerCooldown = 12 * time.Hour
	// 每个接收人每天最多被提醒 5 次
	remindDailyLimit = 5
)

// canRemind：能不能给这个人发短信 / 邮件提醒。
// 业务规则：他本人开了提醒开关 && 手机、邮箱至少绑了一个。
//
// U7 的 can_remind 和 M4 的 can_remind 用的是同一条规则，所以抽出来共用 ——
// 以后要加"黑名单 / 免打扰时段"，只改这一处，两个接口自动跟着变
func canRemind(u *model.User) bool {
	return u.AllowRemind && (u.Phone != "" || u.Email != "")
}

// resolveRemind：M3 提醒的全部决策。
// 顺序有讲究：先判"该不该发"，再看"发得出去吗"，最后才动通道 ——
// 否则会为一个注定不发的提醒白跑两次 COUNT 查询。
//
// ⚠️ 这个函数【故意不返回 error】：文档要求"提醒失败不影响私信本身"，
// 所以提醒侧的一切异常都在这里就地消化成 failed / provider_error，
// 绝不能把错误抛上去，让整条私信跟着一起失败
func resolveRemind(senderID uint, receiver *model.User) model.RemindResult {
	// 两个小工厂：让下面的分支一眼能看出"是跳过还是失败"
	skip := func(reason string) model.RemindResult {
		return model.RemindResult{Status: model.RemindStatusSkipped, Reason: reason}
	}
	fail := func(channel string) model.RemindResult {
		return model.RemindResult{Status: model.RemindStatusFailed, Channel: channel, Reason: model.RemindReasonProviderErr}
	}

	// ① 对方自己把提醒开关关了
	if !receiver.AllowRemind {
		return skip(model.RemindReasonDisabled)
	}
	// ② 手机、邮箱都没绑 → 没有可送达的地址
	if receiver.Phone == "" && receiver.Email == "" {
		return skip(model.RemindReasonNoContact)
	}

	// ③ 限流一：12 小时内我已经提醒过他一次了
	since := time.Now().Add(-remindPeerCooldown)
	samePair, err := repository.CountRemindedTo(senderID, receiver.ID, since)
	if err != nil {
		// 查不出限流状态时【宁可漏发也不多发】：直接判失败，绝不放行
		fmt.Println("❌ 提醒限流查询失败（对端维度）:", err)
		return fail("")
	}
	if samePair > 0 {
		return skip(model.RemindReasonRateLimited)
	}

	// ④ 限流二：他今天已经被提醒 5 次了
	sentToday, err := repository.CountRemindedForReceiver(receiver.ID, startOfToday())
	if err != nil {
		fmt.Println("❌ 提醒限流查询失败（日维度）:", err)
		return fail("")
	}
	if sentToday >= remindDailyLimit {
		return skip(model.RemindReasonRateLimited)
	}

	// ⑤ 真正投递
	channel, err := sendRemindNotice(receiver)
	if err != nil {
		fmt.Println("❌ 提醒投递失败:", err)
		return fail(channel)
	}
	// 成功：reason 留空 —— 枚举里没有"成功"这个 reason
	return model.RemindResult{Status: model.RemindStatusSent, Channel: channel}
}

// startOfToday：今天 0 点（本地时区）。"每天 5 次"从这个时刻开始算，
// 不按"最近 24 小时"—— 否则昨晚 11 点和今早 1 点的两次提醒会被算进同一天
func startOfToday() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
}

// sendRemindNotice：真正把提醒投出去的地方 —— 目前是一根【桩】。
//
// ⚠️ 项目里还没有短信 / 邮件通道（U4「发送验证码」也尚未接入，
// go.mod 里既没有 smtp 也没有任何 SMS SDK），所以这里只打日志、直接返回成功，
// 让前端能按 status=sent 正常联调。
//
// 接入真实服务商时【只改这一个函数】：换成 SDK 调用，失败时返回 error 即可 ——
// resolveRemind 会自动把它变成 failed / provider_error，私信本身照常 201。
//
// 文档硬要求：提醒内容不含私信正文，也不向发送人透露对方联系方式。
// 所以这里既不带正文，也把号码做了脱敏 —— 连日志里都不留明文
func sendRemindNotice(receiver *model.User) (string, error) {
	// 手机优先，其次邮箱（文档原文，不按"哪个先绑"判断）
	if receiver.Phone != "" {
		fmt.Printf("📨 [提醒桩] 手机提醒 → 用户 #%d，号码 %s\n", receiver.ID, maskContact(receiver.Phone))
		return model.RemindChannelSMS, nil
	}
	fmt.Printf("📨 [提醒桩] 邮箱提醒 → 用户 #%d，地址 %s\n", receiver.ID, maskContact(receiver.Email))
	return model.RemindChannelEmail, nil
}

// maskContact：把手机号 / 邮箱打码，只留头尾几个字符。
// 日志是"谁能看到什么"的一条边界线，明文联系方式不该落盘
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
