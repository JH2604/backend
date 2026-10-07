package mail

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"
)

// 超时分两种，因为它们是两种不同的"卡住"：
//   - 拨号超时：根本连不上（网络不通、端口被墙）
//   - 读写超时：连上了但不吭声（对方服务器挂着）
//
// 不设超时的话，发信会一直堵着，把 HTTP 请求也一起拖死。
const (
	dialTimeout = 5 * time.Second
	ioTimeout   = 10 * time.Second
)

// smtpSender 实现 Sender 接口，往真实的 SMTP 服务器投递邮件。
// 它和 mail.go 里的 logSender 是同一接口的两个实现 —— 调用方（service 层）
// 只认 Sender 接口，不知道到底是谁在发，这就是"依赖倒置"。
type smtpSender struct {
	host string // 如 smtp.qq.com
	port string // 如 587
	user string // 发件邮箱，同时也是登录账号（QQ/163 要求 from 必须等于它，不许伪造发件人）
	pass string // 授权码，不是登录密码
}

// NewSMTPSender 造一个"真发信"的发送器。
// 返回接口类型 Sender 而不是具体类型，让调用方只依赖抽象。
func NewSMTPSender(host, port, user, pass string) Sender {
	return smtpSender{host: host, port: port, user: user, pass: pass}
}

// SendVerificationCode 发验证码邮件 —— 主题和正文在这里拼
func (s smtpSender) SendVerificationCode(to, code string) error {
	subject := "【校园失物招领】验证码"
	body := fmt.Sprintf("你的验证码是：%s\r\n\r\n10 分钟内有效，请勿泄露给他人。", code)
	return s.send(to, subject, body)
}

// SendNotification 发通用通知邮件（私信提醒等）—— 主题正文由调用方给
func (s smtpSender) SendNotification(to, subject, body string) error {
	return s.send(to, subject, body)
}

// send 是【唯一的发信实现】。
// 两个公开方法只差"主题正文怎么拼"，其余（拨号、加密、认证、投递）完全一样，
// 所以收在这一个地方 —— 以后要改超时或者换加密方式，只改这里，不会漏。
func (s smtpSender) send(to, subject, body string) error {
	// ① 自己拨号，不用 smtp.SendMail —— 因为标准库那个内部没有超时，会卡死
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(s.host, s.port), dialTimeout)
	if err != nil {
		return fmt.Errorf("连接 SMTP 服务器失败: %w", err)
	}
	defer conn.Close()
	// 整个会话的读写都要在时限内完成
	_ = conn.SetDeadline(time.Now().Add(ioTimeout))

	c, err := smtp.NewClient(conn, s.host)
	if err != nil {
		return fmt.Errorf("SMTP 握手失败: %w", err)
	}
	defer c.Close()

	// ② 587 端口走 STARTTLS：先明文连上，再"就地升级"成加密连接。
	//    （465 端口是隐式 TLS，必须一上来就 tls.Dial，是另一条路；我们用 587 这条常规路）
	if ok, _ := c.Extension("STARTTLS"); ok {
		if err = c.StartTLS(&tls.Config{ServerName: s.host}); err != nil {
			return fmt.Errorf("STARTTLS 升级失败: %w", err)
		}
	}

	// ③ 登录。PlainAuth 会把密码明文发给服务器，所以它【拒绝】在未加密的连接上工作 ——
	//    这正是上一步必须先升级 TLS 的原因（阿里云封 25 端口，也是因为明文太危险）
	if err = c.Auth(smtp.PlainAuth("", s.user, s.pass, s.host)); err != nil {
		return fmt.Errorf("SMTP 认证失败（多半是授权码填错了）: %w", err)
	}

	// ④ 信封：MAIL FROM / RCPT TO
	if err = c.Mail(s.user); err != nil {
		return err
	}
	if err = c.Rcpt(to); err != nil {
		return err
	}

	// ⑤ 投递正文，然后 QUIT 正常收尾
	w, err := c.Data()
	if err != nil {
		return err
	}
	if _, err = w.Write(buildMessage(s.user, to, subject, body)); err != nil {
		return err
	}
	if err = w.Close(); err != nil {
		return err
	}
	return c.Quit()
}

// buildMessage 拼出符合 RFC 5322 的邮件原文。
//
// 关键区别：邮件头分两类，编码规则【不一样】，混了就发不出去。
//   - Subject 这类"非结构化"字段 → 整个值都可以用 RFC 2047 的 =?UTF-8?B?xxx?=
//   - From / To 里的邮箱地址是"结构化"的 addr-spec → 【必须是纯 ASCII 原文】，
//     整个编成 =?...?= 会让服务器读不出地址，QQ 直接回
//     550 "The From header is missing or invalid"（踩过）
//   - 正文用 RFC 2045 的 base64，并在头里声明 Content-Transfer-Encoding
func buildMessage(from, to, subject, body string) []byte {
	var b strings.Builder
	b.WriteString("From: " + from + "\r\n") // ← 地址不编码
	b.WriteString("To: " + to + "\r\n")     // ← 地址不编码
	b.WriteString("Subject: " + encodeHeader(subject) + "\r\n")
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	b.WriteString("Content-Transfer-Encoding: base64\r\n")
	b.WriteString("\r\n") // 空行：头和正文的分界
	b.WriteString(wrapBase64(body))
	return []byte(b.String())
}

// encodeHeader 把含中文的邮件头编成 =?UTF-8?B?xxxx?=
// 【只用于 Subject 这类非结构化字段】—— 千万别拿它编地址。
func encodeHeader(s string) string {
	return "=?UTF-8?B?" + base64.StdEncoding.EncodeToString([]byte(s)) + "?="
}

// wrapBase64 base64 之后的正文每行不能超过 76 字符（RFC 2045 的规定），
// 一整坨超长行有些服务器会直接拒收。
func wrapBase64(s string) string {
	enc := base64.StdEncoding.EncodeToString([]byte(s))
	var b strings.Builder
	for len(enc) > 76 {
		b.WriteString(enc[:76])
		b.WriteString("\r\n")
		enc = enc[76:]
	}
	b.WriteString(enc)
	b.WriteString("\r\n")
	return b.String()
}
