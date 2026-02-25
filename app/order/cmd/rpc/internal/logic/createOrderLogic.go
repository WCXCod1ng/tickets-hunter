package logic

import (
	"context"
	"tickets-hunter/app/order/cmd/rpc/internal/svc"
	"tickets-hunter/app/order/cmd/rpc/order/rpc"
	"tickets-hunter/app/order/model/order_main"
	ticketRpc "tickets-hunter/app/ticket/cmd/rpc/ticket/rpc"
	"tickets-hunter/common/xerr"
	"time"

	errors2 "github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CreateOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateOrderLogic {
	return &CreateOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 创建抢票订单 (包含调用 Ticket RPC 锁座逻辑)
func (l *CreateOrderLogic) CreateOrder(in *rpc.CreateOrderReq) (*rpc.CreateOrderResp, error) {
	// 查询座位信息，防止前端篡改价格等信息
	seatInfo, err := l.svcCtx.TicketRpc.GetSeatInfo(l.ctx, &ticketRpc.GetSeatInfoReq{
		SeatId: in.SeatId,
	})
	if err != nil {
		return nil, err
	}

	// 锁定座位
	lockSeatResp, err := l.svcCtx.TicketRpc.LockSeat(l.ctx, &ticketRpc.LockSeatReq{
		SeatId:  seatInfo.Id,
		EventId: in.EventId,
	})
	if err != nil {
		return nil, errors2.WithStack(err)
	}
	if !lockSeatResp.Success {
		return nil, status.Error(codes.Code(xerr.LockSeatFailed), "锁定座位失败，可能已被其他用户锁定")
	}

	// 到此处座位已成功锁定，后续可以继续执行创单逻辑（如写订单数据库、发送消息等）
	orderSn := l.svcCtx.Snowflake.Generate().String()
	orderMain := &order_main.OrderMain{
		//Id:         0,
		OrderSn:    orderSn,
		UserId:     in.UserId,
		EventId:    in.EventId,
		SeatId:     seatInfo.Id,
		Amount:     seatInfo.Price,
		Status:     OrderStatusPending,
		ExpireTime: time.Now().Add(ExpireDuration),
		//PayTime:    sql.NullTime{},
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
	}
	result, err := l.svcCtx.OrderMainModel.Insert(l.ctx, orderMain)
	// TODO : 这里如果创单失败，应该调用 Ticket RPC 解锁座位，避免死锁，在后续阶段会通过DTM分布式事务框架来保证数据一致性
	if err != nil {
		return nil, errors2.WithStack(status.Error(codes.Internal, err.Error()))
	}
	if rowsAffected, err := result.RowsAffected(); err != nil || rowsAffected == 0 {
		return nil, errors2.WithStack(status.Error(codes.Internal, "创建订单失败，内部错误"))
	}

	return &rpc.CreateOrderResp{
		OrderSn:    orderMain.OrderSn,
		Amount:     orderMain.Amount,
		ExpireTime: orderMain.ExpireTime.Unix(),
	}, nil
}
