package error_pkg

type CodeError struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func (e *CodeError) Error() string {
	return e.Msg
}

func NewCodeError(code int, msg string) error {
	return &CodeError{Code: code, Msg: msg}
}

// 定义常用错误码
const (
	OK                 = 0
	ServerCommonError  = 100001
	RequestParamError  = 100002
	TokenExpireError   = 100003
	TokenGenerateError = 100004
	DbError            = 100005
)
