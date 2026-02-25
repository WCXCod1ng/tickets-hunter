// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package order

import (
	"github.com/zeromicro/go-zero/rest/httpx"
	"net/http"
	"tickets-hunter/app/order/cmd/api/internal/logic/order"
	"tickets-hunter/app/order/cmd/api/internal/svc"
	"tickets-hunter/app/order/cmd/api/internal/types"
	"tickets-hunter/common/result" // 添加这一行引用
)

func CreateOrderHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CreateOrderReq
		if err := httpx.Parse(r, &req); err != nil {
			result.HttpResult(r, w, nil, err) // 解析失败也走统一处理
			return
		}

		l := order.NewCreateOrderLogic(r.Context(), svcCtx)
		resp, err := l.CreateOrder(&req)

		// 使用自定义的 HttpResult
		result.HttpResult(r, w, resp, err)
	}
}
