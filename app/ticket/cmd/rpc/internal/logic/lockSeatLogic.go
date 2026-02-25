package logic

import (
	"context"
	"tickets-hunter/app/ticket/cmd/rpc/internal/svc"
	"tickets-hunter/app/ticket/cmd/rpc/ticket/rpc"

	errors2 "github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	SeatStatusAvailable = 0 // 可售
	SeatStatusLocked    = 1 // 锁定（已被订单锁定，等待支付结果）
	SeatStatusSold      = 2 // 已售（支付成功后更新为已售）
)

type LockSeatLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLockSeatLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LockSeatLogic {
	return &LockSeatLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 锁定座位 (内部隐藏接口，不暴露给前端，仅供 Order RPC 调用)
func (l *LockSeatLogic) LockSeat(in *rpc.LockSeatReq) (*rpc.LockSeatResp, error) {
	success, err := l.svcCtx.TicketSeatModel.UpdateStatusByIdAndStatus(l.ctx, in.SeatId, SeatStatusAvailable, SeatStatusLocked)
	if err != nil {
		return nil, errors2.WithStack(status.Error(codes.Internal, err.Error()))
	}

	return &rpc.LockSeatResp{
		Success: success,
	}, nil
}
