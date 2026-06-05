// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package handler

import (
	"net/http"

	"ai_interview/interview_api/internal/logic"
	"ai_interview/interview_api/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 获取面试历史列表
func ListSessionsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewListSessionsLogic(r.Context(), svcCtx)
		resp, err := l.ListSessions()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
