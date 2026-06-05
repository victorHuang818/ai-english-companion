package error_pkg

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
)

type Body struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

func Response(r *http.Request, w http.ResponseWriter, resp interface{}, err error) {
	if err == nil {
		// 成功返回
		res := &Body{
			Code: OK,
			Msg:  "OK",
			Data: resp,
		}
		httpx.WriteJson(w, http.StatusOK, res)
	} else {
		// 错误返回
		errCode := ServerCommonError
		errMsg := "服务器开小差啦，稍后再来试一试"

		if e, ok := err.(*CodeError); ok {
			// 自定义错误
			errCode = e.Code
			errMsg = e.Msg
		}

		httpx.WriteJson(w, http.StatusOK, &Body{
			Code: errCode,
			Msg:  errMsg,
		})
	}
}
