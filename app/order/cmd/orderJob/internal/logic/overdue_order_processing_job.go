package logic

import (
	"context"
	"tickets-hunter/app/model/order_main"
	"tickets-hunter/app/order/cmd/orderJob/internal/svc"
	"tickets-hunter/app/ticket/cmd/rpc/ticket/rpc"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

type OverdueOrderProcessingJob struct {
	svcCtx *svc.ServiceContext
	Logger logx.Logger
}

func NewOverdueOrderProcessingJob(svcCtx *svc.ServiceContext) *OverdueOrderProcessingJob {
	return &OverdueOrderProcessingJob{
		svcCtx: svcCtx,
		Logger: logx.WithContext(context.Background()),
	}
}

func (j *OverdueOrderProcessingJob) Name() string {
	return "OverdueOrderProcessingJob"
}

func (j *OverdueOrderProcessingJob) Run(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	j.Logger.Infof("%s Run overdue order processing job", j.Name())

	j.processOverdueOrder(ctx)

	for {
		select {
		case <-ctx.Done():
			j.Logger.Infof("任务被外界退出")
			return
		case <-ticker.C:
			j.Logger.Debugf("Ticker触发，执行定时任务")
			j.processOverdueOrder(ctx)
		}
	}
}

// 防止订单没有进入延迟队列（或延迟队列的消息丢失的兜底）
func (j *OverdueOrderProcessingJob) processOverdueOrder(ctx context.Context) {
	// 检查所有超过支付时间但是状态还是待支付的订单
	orders, err := j.svcCtx.OrderMainModel.FindByStatusAndExpireTimeLessThan(ctx, order_main.OrderStatusPending, time.Now())
	if err != nil {
		j.Logger.Errorf("数据库查询错误：%+v", err)
	}
	for _, order := range orders {
		req := &rpc.ReleaseSeatReq{
			SeatId:    order.SeatId,
			OrderSn:   order.OrderSn,
			SeatIndex: order.SeatIndex,
			Section:   order.Section,
			EventId:   order.EventId,
		}
		j.Logger.Infof(req.OrderSn)
		// TODO 优化，批量更新
		// 修改订单表
		success, err := j.svcCtx.OrderMainModel.UpdateStatusByOrderSnAndStatus(ctx, order.OrderSn, order_main.OrderStatusPending, order_main.OrderStatusClosedByTimeout)
		if err != nil || !success {
			j.Logger.Errorf("订单超时兜底任务修改订单表错误，err = %+v, success = %+v", err, success)
		}
		// 调用兜底释放座位的RPC
		resp, err := j.svcCtx.TicketRpc.UnderwriteReleaseSeat(ctx, req)
		if err != nil || resp == nil || !resp.Success {
			j.Logger.Errorf("订单超时兜底任务调用UnderwriteReleaseSeat失败，err = %+v, resp = %+v", err, resp)
		}
		// 可以适当暂停
		time.Sleep(100 * time.Millisecond)
	}
}
