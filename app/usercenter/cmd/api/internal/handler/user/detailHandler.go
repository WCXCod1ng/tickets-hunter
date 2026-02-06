package user

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"tickets-hunter/app/usercenter/cmd/api/internal/logic/user"
	"tickets-hunter/app/usercenter/cmd/api/internal/svc"
)

func DetailHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := user.NewDetailLogic(r.Context(), svcCtx)
		resp, err := l.Detail()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
