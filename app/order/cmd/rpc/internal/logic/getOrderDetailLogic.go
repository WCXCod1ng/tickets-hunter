package logic

import (
	"context"
	"errors"
	"tickets-hunter/app/model/order_main"
	"tickets-hunter/app/order/cmd/rpc/internal/svc"
	"tickets-hunter/app/order/cmd/rpc/order/rpc"

	errors2 "github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GetOrderDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetOrderDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOrderDetailLogic {
	return &GetOrderDetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取订单详情
func (l *GetOrderDetailLogic) GetOrderDetail(in *rpc.GetOrderDetailReq) (*rpc.GetOrderDetailResp, error) {
	userId := in.UserId
	orderSn := in.OrderSn

	order, err := l.svcCtx.OrderMainModel.FindByOrderSnAndUserId(l.ctx, orderSn, userId)
	if errors.Is(err, order_main.ErrNotFound) {
		return nil, status.Error(codes.NotFound, "订单不存在")
	} else if err != nil {
		return nil, errors2.WithStack(status.Error(codes.Internal, err.Error()))
	}

	return &rpc.GetOrderDetailResp{
		OrderSn:    order.OrderSn,
		EventId:    order.EventId,
		SeatId:     order.SeatId,
		Amount:     order.Amount,
		Status:     order.Status,
		ExpireTime: order.ExpireTime.Unix(),
		CreateTime: order.CreateTime.Unix(),
	}, nil
}
