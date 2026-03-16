// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package order

import (
	"context"
	"tickets-hunter/app/order/cmd/rpc/order/rpc"
	"tickets-hunter/common/utils"
	"time"

	"tickets-hunter/app/order/cmd/api/internal/svc"
	"tickets-hunter/app/order/cmd/api/internal/types"

	errors2 "github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CreateOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 创建抢票订单(锁座+创单)
func NewCreateOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateOrderLogic {
	return &CreateOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateOrderLogic) CreateOrder(req *types.CreateOrderReq) (resp *types.CreateOrderResp, err error) {
	// 1. 从上下文中获取用户ID
	userId, err := utils.GetUserIdFromToken(l.ctx)
	if err != nil {
		return nil, errors2.WithStack(status.Error(codes.Unauthenticated, "用户未登录"))
	}

	// 2. 调用 RPC 层创建订单
	rpcResp, err := l.svcCtx.OrderRpc.CreateOrder(l.ctx, &rpc.CreateOrderReq{
		UserId:  userId,
		EventId: req.EventId,
		SeatId:  req.SeatId,
	})
	if err != nil {
		return nil, err
	}

	// 时间戳格式化
	t := time.Unix(rpcResp.ExpireTime, 0)
	expireTimeStr := t.Format("2006-01-02 15:04:05")

	return &types.CreateOrderResp{
		OrderSn:    rpcResp.OrderSn,
		Amount:     float64(rpcResp.Amount) / 100, // 转换为元
		ExpireTime: expireTimeStr,
	}, nil
}
