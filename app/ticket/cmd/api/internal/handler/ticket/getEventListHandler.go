// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package ticket

import (
	"net/http"
	"tickets-hunter/app/ticket/cmd/api/internal/logic/ticket"
	"tickets-hunter/app/ticket/cmd/api/internal/svc"
	"tickets-hunter/common/result" // 添加这一行引用
)

func GetEventListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := ticket.NewGetEventListLogic(r.Context(), svcCtx)
		resp, err := l.GetEventList()

		// 使用自定义的 HttpResult
		result.HttpResult(r, w, resp, err)
	}
}
