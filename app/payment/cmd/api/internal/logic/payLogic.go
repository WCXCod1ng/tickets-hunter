// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"
	"fmt"
	"tickets-hunter/app/model/order_main"
	order_rpc "tickets-hunter/app/order/cmd/rpc/order/rpc"
	"tickets-hunter/app/payment/cmd/rpc/payment"
	"tickets-hunter/common/utils"

	"tickets-hunter/app/payment/cmd/api/internal/svc"
	"tickets-hunter/app/payment/cmd/api/internal/types"

	"github.com/dtm-labs/client/dtmgrpc"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type PayLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 发起订单支付 (Saga 分布式事务)
func NewPayLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PayLogic {
	return &PayLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PayLogic) Pay(req *types.PayReq) (resp *types.PayResp, err error) {
	// 1. 从 JWT Token 中安全解析出 user_id (防越权)
	userId, err := utils.GetUserIdFromToken(l.ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	// 2. 查询订单详情 (金额、座位ID、当前状态)
	// 真实项目中，这里通常是查数据库或调 OrderRPC 的 GetOrder 接口
	orderInfo, err := l.svcCtx.OrderModelMain.FindOneByOrderSn(l.ctx, req.OrderSn)
	if err != nil {
		return nil, err
	}
	if orderInfo.UserId != userId {
		return nil, fmt.Errorf("非法操作：只能支付自己的订单")
	}
	if orderInfo.Status != order_main.OrderStatusPending {
		return nil, fmt.Errorf("订单状态非待支付，无法发起支付")
	}

	// 3. 准备 DTM Saga 事务
	// DTM Server 的 gRPC 地址 (你可以写在 yaml 配置里，这里写死做演示)
	dtmServer := "127.0.0.1:36790"

	// 生成全局分布式事务 ID
	gid := dtmgrpc.MustGenGid(dtmServer)

	// 获取两个 RPC 的 Target (如果你用的 Go-Zero 的 zrpc 并且直连，通常是 IP:Port)
	// 注意：如果你用了 Etcd 注册中心，目标格式通常是 "etcd://127.0.0.1:2379/payment.rpc"
	paymentTarget, err := utils.BuildTarget(l.svcCtx.Config.PaymentRpc)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("构建 Payment RPC Target 失败: %v", err).Error())
	}
	paymentTarget = "127.0.0.1:8083"
	orderRpcTarget, err := utils.BuildTarget(l.svcCtx.Config.OrderRpc)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("构建 Order RPC Target 失败: %v", err).Error())
	}
	orderRpcTarget = "127.0.0.1:8082"

	// 4. 构建 Payload 请求体
	// 因为我们在 .proto 中故意把字段定义得一模一样，所以直接构造两个 protobuf message
	payPayload := &payment.SagaPayReq{
		OrderSn:   orderInfo.OrderSn,
		UserId:    userId,
		Amount:    orderInfo.Amount,
		SeatId:    orderInfo.SeatId,
		EventId:   orderInfo.EventId,
		Section:   orderInfo.Section,
		SeatIndex: orderInfo.SeatIndex,
	}

	ticketPayload := &order_rpc.SagaOrderReq{
		OrderSn:   orderInfo.OrderSn,
		UserId:    userId,
		Amount:    orderInfo.Amount,
		SeatId:    orderInfo.SeatId,
		EventId:   orderInfo.EventId,
		Section:   orderInfo.Section,
		SeatIndex: orderInfo.SeatIndex,
	}

	// 5. ⭐️ 编排 Saga 核心工作流
	saga := dtmgrpc.NewSagaGrpc(dtmServer, gid).
		// 【步骤一】 支付微服务：正向扣款，逆向退款
		// 格式：RPC目标/包名.服务名/方法名
		Add(
			paymentTarget+"/payment.Payment/Deduct",
			paymentTarget+"/payment.Payment/Refund",
			payPayload,
		).
		// 【步骤二】 票务微服务：正向出票，逆向回滚退票
		Add(
			orderRpcTarget+"/order.OrderService/IssueTicket",
			orderRpcTarget+"/order.OrderService/RollbackTicket",
			ticketPayload,
		)

	// 6. 开启同步等待模式（重要）
	// 设置 WaitResult 为 true，网关会一直阻塞等待 DTM 走完整个流程（包括失败时的回滚）。
	// 这样可以直接把最终成功/失败的状态返回给前端！
	saga.WaitResult = true
	saga.RequestTimeout = 30 // 设置请求超时时间，单位秒，默认是 20 秒，根据实际情况调整

	// 7. 提交全局事务
	err = saga.Submit()
	if err != nil {
		// DTM 判定事务最终失败（比如余额不足、或出票失败且退款已完成）
		l.Logger.Errorf("订单支付-出票全链路最终失败, order_sn: %s, err: %v", req.OrderSn, err)
		return &types.PayResp{
			Success: false,
			Message: "支付或出票失败，若已扣款将自动退回，请检查订单状态",
		}, nil
	}

	// 8. 全部顺利通关！
	return &types.PayResp{
		Success: true,
		Message: "支付成功，已为您锁定出票！",
	}, nil
	return
}
