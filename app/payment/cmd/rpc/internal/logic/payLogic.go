package logic

import (
	"context"
	"tickets-hunter/app/payment/cmd/rpc/internal/svc"
	"tickets-hunter/app/payment/cmd/rpc/payment/rpc"

	"github.com/zeromicro/go-zero/core/logx"
)

type PayLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPayLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PayLogic {
	return &PayLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// API 层调用的支付接口
func (l *PayLogic) Pay(in *rpc.PayReq) (*rpc.PayResp, error) {

	return &rpc.PayResp{}, nil
}
