// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"net/http"
	"tickets-hunter/app/usercenter/cmd/api/internal/logic/user"
	"tickets-hunter/app/usercenter/cmd/api/internal/svc"
	"tickets-hunter/app/usercenter/cmd/api/internal/types"
	"tickets-hunter/common/result" // 添加这一行引用

	"github.com/zeromicro/go-zero/rest/httpx"
)

func LogoutHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.LogoutReq
		if err := httpx.Parse(r, &req); err != nil {
			result.HttpResult(r, w, nil, err) // 解析失败也走统一处理
			return
		}

		l := user.NewLogoutLogic(r.Context(), svcCtx)
		err := l.Logout(&req)

		// 使用自定义的 HttpResult
		result.HttpResult(r, w, nil, err)
	}
}
