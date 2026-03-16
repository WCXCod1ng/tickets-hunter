package logic

import (
	"context"

	"tickets-hunter/app/ticket/cmd/rpc/internal/svc"
	"tickets-hunter/app/ticket/cmd/rpc/ticket/rpc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetSeatBitMapLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetSeatBitMapLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSeatBitMapLogic {
	return &GetSeatBitMapLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取BitMap信息
func (l *GetSeatBitMapLogic) GetSeatBitMap(in *rpc.GetSeatBitMapReq) (*rpc.GetEventListResp, error) {
	// todo: add your logic here and delete this line

	return &rpc.GetEventListResp{}, nil
}
