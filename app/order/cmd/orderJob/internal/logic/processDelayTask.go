package logic

// 处理延时任务，在当前的实现中，主要是处理订单超时未支付的情况，执行释放座位等相关逻辑
// 该逻辑会在 orderJob 服务中以独立的 goroutine 方式运行，持续从 Redis 延时队列拉取超时订单并处理

import (
	"context"
	"tickets-hunter/app/model/order_main"
	"tickets-hunter/app/ticket/cmd/rpc/ticket/rpc"
	"time"

	"tickets-hunter/app/order/cmd/orderJob/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/threading"
)

type ProcessDelayTaskJob struct {
	svcCtx *svc.ServiceContext
	Logger logx.Logger
}

func NewProcessDelayTaskJob(svcCtx *svc.ServiceContext) *ProcessDelayTaskJob {
	return &ProcessDelayTaskJob{
		svcCtx: svcCtx,
		Logger: logx.WithContext(context.Background()),
	}
}

func (j *ProcessDelayTaskJob) Name() string {
	return "ProcessDelayTaskJob"
}

func (j *ProcessDelayTaskJob) Run(ctx context.Context) {
	j.Logger.Infof("延迟队列消费者已启动...")

	for {
		// 处理上下文取消，优雅退出
		select {
		case <-ctx.Done():
			j.Logger.Infof("延迟队列消费者收到退出信号，正在关闭...")
			return
		default:
		}

		// 获取batchSize个超时订单
		batchSize := 100 // 可以根据实际业务需求调整批量处理的数量，过大可能导致单次处理时间过长，过小则可能增加系统负载
		orderSns, err := j.svcCtx.DelayQueue.GetDelayTaskBatch(ctx, batchSize)
		if err != nil { // 出现网络异常时，休眠防死循环打满 CPU
			j.Logger.Errorf("拉取延迟队列异常: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}
		if len(orderSns) == 0 {
			// 没有超时订单，睡眠等待一段时间后继续拉取，避免空轮询导致的资源浪费
			time.Sleep(4 * time.Second)
			continue
		}

		j.Logger.Debugf("当前有超时订单，进行处理 orderSns: %v", orderSns)
		// 释放订单对应的座位等相关资源
		j.processDelayedOrder(ctx, orderSns)

		// 🚀 重点优化：处理完毕后，不再固定休眠 5 秒！
		// 如果本次拉取到的数量 == batchSize，说明队列里大概率还有积压的货，立刻进入下一次循环(不休眠)。
		// 如果本次拉取的数量 < batchSize，说明队列已经快被抽干了，稍微喘口气(防空转)。
		if len(orderSns) < batchSize {
			time.Sleep(500 * time.Millisecond) // 极短的休眠
		}
	}
}

// 处理超时订单的逻辑，主要包括：
// 1. 查 MySQL 获取订单详情
// 2. 状态机幂等检查，确保订单状态仍然是待支付，防止用户已经支付成功但延迟队列才把订单弹出来的情况
// 3. 修改 MySQL 订单状态 从 10 (待支付) 改为 40 (已取消)，并且要采用乐观锁机制，确保只有在订单状态未被修改的情况下才更新成功，避免并发修改导致的数据不一致问题
// 4. 调用 Ticket RPC 释放 Redis 锁并且恢复 MySQL 座位状态
func (j *ProcessDelayTaskJob) processDelayedOrder(ctx context.Context, orderSns []string) {
	runner := threading.NewTaskRunner(20)

	for _, os := range orderSns {
		// 🚀 开启 Goroutine 异步处理单条订单
		orderSn := os
		runner.Schedule(func() {

			// 为了防止协程内的 panic 导致整个进程崩溃，工业界通常会加上 recover
			defer func() {
				if r := recover(); r != nil {
					j.Logger.Errorf("处理订单 %s 发生 Panic: %v", orderSn, r)
				}
			}()

			// ==========================================
			// 1. 查 MySQL 获取订单详情
			// ==========================================
			orderRecord, err := j.svcCtx.OrderMainModel.FindOneByOrderSn(ctx, orderSn)
			if err != nil || orderRecord == nil {
				j.Logger.Errorf("[延迟队列] 超时订单 %s 在数据库中不存在或查询失败: %v", orderSn, err)
				return
			}

			// ==========================================
			// 2. 状态机幂等检查 (核心防御！)
			// ==========================================
			// 假设用户在第 14 分 59 秒支付成功，状态已经被改为了 20 (已支付)。
			// 此时延迟队列刚好把这个订单弹出来，如果不做检查直接取消，就会造成灾难（收了钱却把票退了）。
			if orderRecord.Status != order_main.OrderStatusPending { // 10 代表"待支付"
				j.Logger.Debugf("[延迟队列] 订单 %s 状态不为待支付(当前状态:%d)，说明已支付或已处理，跳过释放", orderSn, orderRecord.Status)
				return
			}

			// ==========================================
			// 3. 修改 MySQL 订单状态 从 10 (待支付) 改为 40 (已取消)，并且要采用乐观锁机制，确保只有在订单状态未被修改的情况下才更新成功，避免并发修改导致的数据不一致问题
			// ==========================================
			// UPDATE order_main SET status = 40 WHERE order_sn = ? AND status = 10
			success, err := j.svcCtx.OrderMainModel.UpdateStatusByOrderSnAndStatus(ctx, orderSn, order_main.OrderStatusPending, order_main.OrderStatusClosedByTimeout)
			if err != nil || !success {
				j.Logger.Errorf("[延迟队列] 修改订单 %s 状态为 40 失败，可能发生了并发修改", orderSn)
				return
			}

			j.Logger.Debugf("[延迟队列] 订单 %s 已成功标记为超时关闭", orderSn)

			// ==========================================
			// 4. 调用 Ticket RPC 释放 Redis 锁并且恢复 MySQL 座位状态
			// ==========================================
			_, err = j.svcCtx.TicketRpc.ReleaseSeat(ctx, &rpc.ReleaseSeatReq{
				SeatId:    orderRecord.SeatId,
				OrderSn:   orderRecord.OrderSn,
				SeatIndex: orderRecord.SeatIndex,
				Section:   orderRecord.Section,
				EventId:   orderRecord.EventId,
			})
			if err != nil {
				j.Logger.Errorf("[延迟队列] 通知 Ticket RPC 释放 MySQL 座位状态失败, SeatId: %d, err: %v", orderRecord.SeatId, err)
				// TODO 通过补偿，而非分布式事务
			} else {
				j.Logger.Debugf("[延迟队列] 座位 %d 已彻底释放回票池", orderRecord.SeatId)
			}
		})
	}

	runner.Wait()
}
