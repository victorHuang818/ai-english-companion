// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package handler

import (
	"net/http"

	"ai_companion/companion_api/internal/logic"
	"ai_companion/companion_api/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 获取当前用户的档案列表
func ListUserProfilesHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewListUserProfilesLogic(r.Context(), svcCtx)
		resp, err := l.ListUserProfiles()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
