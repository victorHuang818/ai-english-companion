// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package handler

import (
	"net/http"

	"ai_companion/companion_api/internal/logic"
	"ai_companion/companion_api/internal/svc"
	"ai_companion/companion_api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 获取面试详情
func GetSessionDetailHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetSessionDetailReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewGetSessionDetailLogic(r.Context(), svcCtx)
		resp, err := l.GetSessionDetail(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
