package mail

import (
	"fmt"
)

type Sender interface {
	SendVerificationCode(to, code string) error
	SendNotification(to, subject, body string) error
}

var Default Sender = logSender{}

type logSender struct{}

func (logSender) SendVerificationCode(to, code string) error {
	fmt.Printf("📧 [DEV] 验证码 → %s : %s\n", to, code)

	return nil

}

func (logSender) SendNotification(to, subject, body string) error {
	fmt.Printf("📧 [DEV] 通知 → %s\n【%s】\n%s\n", to, subject, body)
	return nil
}
