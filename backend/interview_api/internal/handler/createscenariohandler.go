// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package handler

import (
	"net/http"

	"ai_interview/interview_api/internal/logic"
	"ai_interview/interview_api/internal/svc"
	"ai_interview/interview_api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 创建练习场景
func CreateScenarioHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CreateScenarioReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewCreateScenarioLogic(r.Context(), svcCtx)
		resp, err := l.CreateScenario(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
