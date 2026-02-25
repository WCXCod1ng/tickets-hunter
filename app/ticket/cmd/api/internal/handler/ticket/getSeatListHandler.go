// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package ticket

import (
	"github.com/zeromicro/go-zero/rest/httpx"
	"net/http"
	"tickets-hunter/app/ticket/cmd/api/internal/logic/ticket"
	"tickets-hunter/app/ticket/cmd/api/internal/svc"
	"tickets-hunter/app/ticket/cmd/api/internal/types"
	"tickets-hunter/common/result" // 添加这一行引用
)

func GetSeatListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.SeatListReq
		if err := httpx.Parse(r, &req); err != nil {
			result.HttpResult(r, w, nil, err) // 解析失败也走统一处理
			return
		}

		l := ticket.NewGetSeatListLogic(r.Context(), svcCtx)
		resp, err := l.GetSeatList(&req)

		// 使用自定义的 HttpResult
		result.HttpResult(r, w, resp, err)
	}
}
