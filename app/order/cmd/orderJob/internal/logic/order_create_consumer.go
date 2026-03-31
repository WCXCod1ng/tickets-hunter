package logic

import (
	"context"
	"database/sql"
	"tickets-hunter/app/model/order_main"
	"tickets-hunter/app/model/ticket_seat"
	"tickets-hunter/app/ticket/cmd/rpc/ticket/rpc"
	"tickets-hunter/common/mq"
	"tickets-hunter/common/msg/order_msg"
	"time"

	"tickets-hunter/app/order/cmd/orderJob/internal/svc"

	"github.com/zeromicro/go-queue/kq"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/queue"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"google.golang.org/protobuf/proto"
)

type OrderCreateConsumerJob struct {
	svcCtx *svc.ServiceContext
	Logger logx.Logger
	MQ     queue.MessageQueue
}

func NewOrderCreateConsumerJob(svcCtx *svc.ServiceContext) *OrderCreateConsumerJob {

	return &OrderCreateConsumerJob{
		svcCtx: svcCtx,
		Logger: logx.WithContext(context.Background()),
	}
}

func (c *OrderCreateConsumerJob) Name() string {
	return "OrderCreateConsumerJob"
}

func (c *OrderCreateConsumerJob) Run(ctx context.Context) {
	orderDispatcher := mq.NewDispatcher()
	orderDispatcher.Register("order_create", func(ctx context.Context, msg []byte) error {
		return c.consume(ctx, msg)
	})

	orderConsumer := mq.NewConsumer(orderDispatcher)
	orderConsumerAdapter := mq.NewMQConsumerAdapter(orderConsumer)
	c.MQ = kq.MustNewQueue(c.svcCtx.Config.Kafka, orderConsumerAdapter)
	c.Logger.Infof("创建订单消费者已启动")
	c.MQ.Start()
}

func (c *OrderCreateConsumerJob) consume(ctx context.Context, data []byte) error {

	c.Logger.Infof("接收到消息")

	var msg order_msg.CreateOrderMsg
	if err := proto.Unmarshal(data, &msg); err != nil {
		logx.Errorf("Unmarshal Kafka msg failed: %v", err)
		return nil // 消息格式错误，无法重试，直接返回 nil 跳过该消息
	}

	// ==========================================================
	// 1. 【新增】幂等性检查（防止 Kafka 消息重复投递导致误报错误）
	// ==========================================================
	existingOrder, err := c.svcCtx.OrderMainModel.FindOneByOrderSn(ctx, msg.OrderSn)
	if err != nil && err != sqlx.ErrNotFound {
		c.Logger.Errorf("查询订单失败，引发重试: %v", err)
		return err // 数据库宕机等系统异常，return err 让 Kafka 稍后重试
	}
	if existingOrder != nil {
		c.Logger.Infof("订单 %s 已存在，忽略重复投递的消息", msg.OrderSn)
		return nil // 已经处理过了，直接返回成功，丢弃消息
	}

	// ==========================================================
	// 2. 更新座位表状态
	// ==========================================================
	success, err := c.svcCtx.TicketSeatModel.UpdateStatusByIdAndOldStatus(ctx, msg.SeatId, ticket_seat.SeatStatusAvailable, ticket_seat.SeatStatusLocked)
	if err != nil {
		c.Logger.Errorf("数据库异常导致更新座位失败，引发重试: %v", err)
		return err // 系统异常（如断网），此时 MySQL 还没脏数据，return err 让 Kafka 重试
	}
	if !success {
		// 业务异常：数据库没报错，但是 affected_rows = 0。说明座位状态已经被别人改了（或者脏数据）
		c.Logger.Errorf("座位状态不为Available，抢票失败, order_sn: %s", msg.OrderSn)

		// 补偿 1：只需要释放 Redis 锁（因为 MySQL 根本没修改成功）
		c.rollbackRedisLock(ctx, msg.SeatId, msg.OrderSn, msg.SeatIndex, msg.Section, msg.EventId)
		// 补偿 2：通知前端抢票失败（写 Redis 状态）
		c.notifyFrontend(ctx, msg.OrderSn, "failed")

		return nil // 业务规则导致的失败，重试没用，直接 return nil 结束
	}

	// ==========================================================
	// 3. 插入订单表
	// ==========================================================
	orderMain := &order_main.OrderMain{
		Id:         0,
		OrderSn:    msg.OrderSn,
		UserId:     msg.UserId,
		EventId:    msg.EventId,
		SeatId:     msg.SeatId,
		Section:    msg.Section,
		SeatIndex:  msg.SeatIndex,
		Amount:     msg.Amount,
		Status:     order_main.OrderStatusPending,
		ExpireTime: time.Unix(msg.ExpireTime, 0),
		PayTime:    sql.NullTime{},
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
	}

	_, err = c.svcCtx.OrderMainModel.Insert(ctx, orderMain)
	if err != nil {
		c.Logger.Errorf("插入订单表出错：%+v", err)

		// 此时出现了“部分成功”的脏状态：MySQL的座位表被改成 Locked 了，但订单没插进去！
		// 补偿 1：释放 Redis 锁
		// 补偿 2：把 MySQL 座位表改回 Available
		// （你可以直接调用你原来的 ReleaseSeat RPC，因为它兼顾了修改 MySQL 和 Redis）
		c.rollbackMySQLAndRedis(ctx, msg.SeatId, msg.OrderSn, msg.SeatIndex, msg.Section, msg.EventId)

		c.notifyFrontend(ctx, msg.OrderSn, "failed")
		return nil // 同样，作为简化的异常处理，回滚后直接结束
	}

	// ==========================================================
	// 4. 【修改】动态计算延迟队列的投递时间
	// ==========================================================
	expireTime := time.Unix(msg.ExpireTime, 0)
	delayDuration := time.Until(expireTime) // 动态计算剩余时间

	if delayDuration > 0 {
		err = c.svcCtx.DelayQueue.Add(ctx, orderMain.OrderSn, delayDuration)
		if err != nil {
			c.Logger.Errorf("订单 %s 投递延迟队列失败: %v", orderMain.OrderSn, err)
			// 投递失败不影响主流程，靠定时任务兜底
		}
	} else {
		c.Logger.Errorf("消息积压严重，处理时已超过订单过期时间 order_sn: %s", msg.OrderSn)
		// 如果你想做得很完善，这里可以直接触发取消订单逻辑
	}

	// ==========================================================
	// 5. 通知前端抢票成功
	// ==========================================================
	c.notifyFrontend(ctx, msg.OrderSn, "success")
	c.Logger.Infof("create order success, order_sn = %s", orderMain.OrderSn)

	return nil // commit offset automatically
}

func (c *OrderCreateConsumerJob) notifyFrontend(ctx context.Context, sn string, s string) {
	c.Logger.Infof("create order %s, order_sn = %s", sn, s)
}

func (c *OrderCreateConsumerJob) rollbackMySQLAndRedis(ctx context.Context, seatId int64, orderSn string, seatIndex int64, section string, eventId int64) {
	req := &rpc.ReleaseSeatReq{
		SeatId:    seatId,
		OrderSn:   orderSn,
		SeatIndex: seatIndex,
		Section:   section,
		EventId:   eventId,
	}

	// 由于ReleaseSeat还是同时释放Redis和MySQL，所以这里直接调用了
	resp, err := c.svcCtx.TicketRpc.ReleaseSeat(ctx, req)
	if err != nil || resp == nil || !resp.Success {
		c.Logger.Errorf("Release Seat error: err = %+v, resp = %+v", err, resp)
	}
}

func (c *OrderCreateConsumerJob) rollbackRedisLock(ctx context.Context, seatId int64, orderSn string, seatIndex int64, section string, eventId int64) {
	req := &rpc.UnlockSeatReq{
		SeatId:    seatId,
		OrderSn:   orderSn,
		SeatIndex: seatIndex,
		Section:   section,
		EventId:   eventId,
	}

	_, err := c.svcCtx.TicketRpc.UnlockSeat(ctx, req)
	if err != nil {
		// 不关心锁释放失败的情况，统一由定时任务兜底
		c.Logger.Errorf("UnlockSeat error: err = %+v", err)
	}
}
