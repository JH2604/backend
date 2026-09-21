package errcode

// 错误码常量：一个 code 对应一个含义
const (
	Success          = 0     // 成功
	ErrUserExists    = 20001 // 用户名已存在
	ErrInvalidParams = 40001 // 参数错误
	ErrNotFound      = 40002 // 资源不存在
	ErrUnauthorized  = 40003 // 未登录或登录已过期
	ErrForbidden     = 40004 // 没有权限
	ErrServer        = 50000 // 服务器内部错误
)

// 码 → 消息 的映射表（作业要求：「一个 code 对应一个 msg」）
var messages = map[int]string{
	Success:          "成功",
	ErrUserExists:    "用户名已存在",
	ErrInvalidParams: "参数错误",
	ErrNotFound:      "资源不存在",
	ErrUnauthorized:  "未登录或登录已过期",
	ErrForbidden:     "没有权限",
	ErrServer:        "服务器内部错误",
}

// GetMsg 根据错误码取对应的消息
func GetMsg(code int) string {
	if msg, ok := messages[code]; ok {
		return msg
	}
	return "未知错误"
}
