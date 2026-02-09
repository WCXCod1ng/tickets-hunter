// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"net/http"
	"tickets-hunter/app/usercenter/cmd/api/internal/logic/user"
	"tickets-hunter/app/usercenter/cmd/api/internal/svc"
	"tickets-hunter/common/result" // 添加这一行引用
)

func DetailHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := user.NewDetailLogic(r.Context(), svcCtx)
		resp, err := l.Detail()

		// 使用自定义的 HttpResult
		result.HttpResult(r, w, resp, err)
	}
}
