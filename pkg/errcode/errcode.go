package errcode

// 错误码常量：一个 code 对应一个含义
const (
	Success             = 0     // 成功
	ErrInvalidParams    = 40000 // 参数错误
	ErrCodeExpired      = 40001 // 验证码错误或过期
	ErrOldPassword      = 40002 // 原密码错误
	ErrNoToken          = 40100 // 未登录或登录已失效
	ErrTokenTimeout     = 40101 // 登录已过期
	ErrBadCredentials   = 40102 // 学号或密码错误
	ErrSessionInvalid   = 40103 // 登录已失效，请重新登录
	ErrRefreshToken     = 40104 // 登录凭证异常，请重新登录
	ErrPermission       = 40300 // 没有权限
	ErrNotFound         = 40400 // 资源不存在
	ErrStudentNotFound  = 40401 // 学号不在实名库中
	ErrResourceConflict = 40900 // 资源冲突
	ErrFileTooLarge     = 41300 // 文件过大
	ErrFileType         = 41500 // 文件类型不支持
	ErrTooManyRequests  = 42900 // 请求过于频繁
	ErrServer           = 50000 // 服务器内部错误

)

// 码 → 消息 的映射表（作业要求：「一个 code 对应一个 msg」）
var messages = map[int]string{
	Success:             "成功",
	ErrInvalidParams:    "参数错误",
	ErrNotFound:         "资源不存在",
	ErrCodeExpired:      "验证码错误或过期",
	ErrOldPassword:      "原密码错误",
	ErrNoToken:          "未登录或登录已失效",
	ErrTokenTimeout:     "登录已过期",
	ErrBadCredentials:   "学号或密码错误",
	ErrSessionInvalid:   "登录已失效，请重新登录",
	ErrRefreshToken:     "登录凭证异常，请重新登录",
	ErrPermission:       "没有权限",
	ErrStudentNotFound:  "学号不在实名库中",
	ErrResourceConflict: "资源冲突",
	ErrFileTooLarge:     "文件过大",
	ErrFileType:         "文件类型不支持",
	ErrTooManyRequests:  "请求过于频繁",
	ErrServer:           "服务器内部错误",
}

// GetMsg 根据错误码取对应的消息
func GetMsg(code int) string {
	if msg, ok := messages[code]; ok {
		return msg
	}
	return "未知错误"
}
