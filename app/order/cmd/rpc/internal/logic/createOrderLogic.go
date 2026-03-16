package logic

import (
	"context"
	"database/sql"
	"tickets-hunter/app/model/order_main"
	"tickets-hunter/app/order/cmd/rpc/internal/svc"
	"tickets-hunter/app/order/cmd/rpc/order/rpc"
	"tickets-hunter/app/order/define"
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
	// 0. 查询座位信息，防止前端篡改价格等信息
	seatInfo, err := l.svcCtx.TicketRpc.GetSeatInfo(l.ctx, &ticketRpc.GetSeatInfoReq{
		SeatId: in.SeatId,
	})
	if err != nil {
		return nil, err
	}

	// 1. 生成订单号
	orderSn := l.svcCtx.Snowflake.Generate().String()

	// 2. 锁定座位
	lockSeatResp, err := l.svcCtx.TicketRpc.LockSeat(l.ctx, &ticketRpc.LockSeatReq{
		EventId:   in.EventId,
		SeatId:    seatInfo.Id,
		OrderSn:   orderSn,
		SeatIndex: seatInfo.SeatIndex,
		Section:   seatInfo.Section,
	})
	if err != nil {
		return nil, errors2.WithStack(err)
	}
	if !lockSeatResp.Success {
		return nil, status.Error(codes.Code(xerr.LockSeatFailed), "锁定座位失败，可能已被其他用户锁定")
	}

	// TODO 引入Kafka来削峰，后续的逻辑不再直接在这里执行，而是投递一个消息到 Kafka，异步消费这个消息来执行后续的创单、投递延迟队列等逻辑，这样可以大大降低下单接口的响应时间，提升用户体验，同时也能更好地应对高并发场景，避免数据库压力过大导致系统崩溃

	// 到此处座位已成功锁定，后续可以继续执行创单逻辑（如写订单数据库、发送消息等）
	// 3. 写入订单
	orderMain := &order_main.OrderMain{
		//Id:         0,
		OrderSn:    orderSn,
		UserId:     in.UserId,
		EventId:    in.EventId,
		SeatId:     seatInfo.Id,
		Section:    seatInfo.Section,
		SeatIndex:  seatInfo.SeatIndex,
		Amount:     seatInfo.Price,
		Status:     order_main.OrderStatusPending,
		ExpireTime: time.Now().Add(define.ExpireDuration),
		PayTime:    sql.NullTime{},
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

	// 4. 投递到延迟队列
	// 设置延迟时间，代表支付超时时间，超过这个时间未支付则自动取消订单并释放锁定的座位
	// 到期后会有一个后台任务不断地从这个队列里拉取超时订单的订单号，执行相关的超时处理逻辑（如修改订单状态、调用 Ticket RPC 释放座位等）
	// debug阶段设置为1分钟，实际生产环境中一般设置为15分钟
	err = l.svcCtx.DelayQueue.Add(l.ctx, orderMain.OrderSn, define.ExpireDuration)
	if err != nil {
		l.Logger.Errorf("订单 %s 投递延迟队列失败: %v", orderSn, err)
		// 注意，这里哪怕投递失败，也不会中断用户下单的主流程
		// 工业界通常会打印Error日志，并配合外部补偿脚本、定时扫表等方式来处理这种边缘情况，确保最终一致性
		return nil, err
	}

	return &rpc.CreateOrderResp{
		OrderSn:    orderMain.OrderSn,
		Amount:     orderMain.Amount,
		ExpireTime: orderMain.ExpireTime.Unix(),
	}, nil
}
