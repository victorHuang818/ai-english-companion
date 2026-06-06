// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package handler

import (
	"net/http"

	"ai_companion/companion_api/internal/logic"
	"ai_companion/companion_api/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 获取场景列表
func ListScenariosHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewListScenariosLogic(r.Context(), svcCtx)
		resp, err := l.ListScenarios()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
