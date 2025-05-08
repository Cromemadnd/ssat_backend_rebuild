package utils

type ErrorCode struct {
	Code     int
	HttpCode int
	Message  string
}

func (e ErrorCode) Error() string {
	return e.Message
}

var (
	ErrOK = ErrorCode{
		Code:     0,
		HttpCode: 200,
		Message:  "请求成功",
	}
	ErrCreated = ErrorCode{
		Code:     0,
		HttpCode: 201,
		Message:  "创建成功",
	}
	ErrNotFound = ErrorCode{
		Code:     1,
		HttpCode: 404,
		Message:  "未找到",
	}
	ErrUnauthorized = ErrorCode{
		Code:     2,
		HttpCode: 401,
		Message:  "未登录",
	}
	ErrInternalServer = ErrorCode{
		Code:     3,
		HttpCode: 500,
		Message:  "服务器内部错误",
	}
	ErrBadRequest = ErrorCode{
		Code:     4,
		HttpCode: 400,
		Message:  "请求错误",
	}
	ErrMissingParam = ErrorCode{
		Code:     5,
		HttpCode: 400,
		Message:  "缺少参数",
	}
	ErrIncorrectAuthInfo = ErrorCode{
		Code:     6,
		HttpCode: 400,
		Message:  "用户名或密码错误",
	}
	ErrUserExists = ErrorCode{
		Code:     7,
		HttpCode: 400,
		Message:  "用户已存在",
	}
	ErrUserNotFound = ErrorCode{
		Code:     8,
		HttpCode: 404,
		Message:  "用户未找到",
	}
	ErrInvalidJWT = ErrorCode{
		Code:     9,
		HttpCode: 404,
		Message:  "认证信息无效",
	}
	ErrExpiredJWT = ErrorCode{
		Code:     10,
		HttpCode: 401,
		Message:  "认证信息已过期，请重新登录",
	}
	ErrAlreadyLoggedIn = ErrorCode{
		Code:     11,
		HttpCode: 400,
		Message:  "已登录",
	}
	ErrUnknownDevice = ErrorCode{
		Code:     12,
		HttpCode: 404,
		Message:  "未知的设备ID",
	}
	ErrInvalidSignature = ErrorCode{
		Code:     13,
		HttpCode: 400,
		Message:  "无效的签名",
	}
	ErrExpiredRequest = ErrorCode{
		Code:     14,
		HttpCode: 400,
		Message:  "时间戳过旧",
	}
	ErrReplayAttack = ErrorCode{
		Code:     15,
		HttpCode: 400,
		Message:  "检测到重放攻击",
	}
	ErrForbidden = ErrorCode{
		Code:     1001,
		HttpCode: 403,
		Message:  "权限不足",
	}
)
