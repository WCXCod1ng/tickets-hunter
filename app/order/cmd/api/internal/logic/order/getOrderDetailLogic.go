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

type GetOrderDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取订单详情
func NewGetOrderDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOrderDetailLogic {
	return &GetOrderDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetOrderDetailLogic) GetOrderDetail(req *types.OrderDetailReq) (resp *types.OrderDetailResp, err error) {
	// 获取用户Id
	userId, err := utils.GetUserIdFromToken(l.ctx)
	if err != nil {
		return nil, errors2.WithStack(status.Error(codes.Unauthenticated, "用户未登录"))
	}

	// 调用 RPC 层获取订单详情
	rpcResp, err := l.svcCtx.OrderRpc.GetOrderDetail(l.ctx, &rpc.GetOrderDetailReq{
		OrderSn: req.OrderSn,
		UserId:  userId,
	})
	if err != nil {
		return nil, err
	}

	// 时间戳格式化
	expireTimeStr := time.Unix(rpcResp.ExpireTime, 0).Format("2006-01-02 15:04:05")
	createTimeStr := time.Unix(rpcResp.CreateTime, 0).Format("2006-01-02 15:04:05")

	return &types.OrderDetailResp{
		OrderSn:    rpcResp.OrderSn,
		EventId:    rpcResp.EventId,
		SeatId:     rpcResp.SeatId,
		Amount:     float64(rpcResp.Amount) / 100, // 转换为元
		Status:     rpcResp.Status,
		ExpireTime: expireTimeStr,
		CreateTime: createTimeStr,
	}, nil
}
