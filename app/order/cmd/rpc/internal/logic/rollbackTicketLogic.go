package logic

import (
	"context"
	"database/sql"
	"tickets-hunter/app/model/order_main"
	"tickets-hunter/app/model/ticket_seat"
	"tickets-hunter/app/order/cmd/rpc/order/rpc"
	redis2 "tickets-hunter/common/redis"

	"tickets-hunter/app/order/cmd/rpc/internal/svc"

	"github.com/dtm-labs/client/dtmgrpc"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type RollbackTicketLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRollbackTicketLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RollbackTicketLogic {
	return &RollbackTicketLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Saga 步骤2 补偿操作：回滚出票
func (l *RollbackTicketLogic) RollbackTicket(in *rpc.SagaOrderReq) (*rpc.SagaOrderResp, error) {
	l.Logger.Debugf("开始执行出票回滚补偿逻辑，orderSn: %s", in.OrderSn)
	barrier, err := dtmgrpc.BarrierFromGrpc(l.ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	db, err := l.svcCtx.DB.RawDB()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	// 数据库层面回滚
	err = barrier.CallWithDB(db, func(tx *sql.Tx) error {
		// 1. 恢复订单状态：改为 51(已退款关闭)
		// 注意条件：只要不是 51，就强行改成 51。防重入和幂等由 barrier 表保障。
		success, err := l.svcCtx.OrderMainModel.UpdateStatusByOrderSnAndNotStatusWithTx(l.ctx, tx, in.OrderSn, order_main.OrderStatusRefunded, order_main.OrderStatusRefunded)
		if err != nil {
			return err
		}

		if !success {
			l.Logger.Debugf("订单状态已是已退款关闭，无需重复回滚，order_sn: %s", in.OrderSn)
		}

		// 2. 恢复座位状态：将 1(锁定) 或极小概率的 2(已售) 强行恢复为 0(可选)
		success, err = l.svcCtx.TicketSeatModel.UpdateStatusByIdAndInStatusWithTx(l.ctx, tx, in.SeatId, []int64{ticket_seat.SeatStatusLocked, ticket_seat.SeatStatusSold}, ticket_seat.SeatStatusAvailable)

		if err != nil {
			return err
		}

		if !success {
			l.Logger.Debugf("座位状态已是可选，无需重复回滚，seat_id: %d", in.SeatId)
		}

		return nil
	})

	if err != nil {
		l.Logger.Errorf("回滚出票失败，DTM 将不断重试. order_sn: %s, err: %v", in.OrderSn, err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Redis 缓存同步
	// 阶段三你 Redis 维护了可用座位池 (valid_seats)，
	// 数据库回滚为 0 后，最好将这个 SeatIndex 重新添加回BitMap，
	// 并删除 ticket:seat:{seatId}:lock 锁，让其他用户能立刻抢到。
	// 使用Lua脚本保证原子性
	keys := []string{
		redis2.SeatLockRedisKey(in.SeatId),
		redis2.SeatBitMapRedisKey(in.EventId, in.Section),
		redis2.OrderDelayQueueKey,
	}
	values := []any{
		in.OrderSn,
		in.SeatIndex,
	}

	// 注意，应当保证Redis操作被执行成功，如果Redis操作失败了，虽然数据库已经回滚了，但Redis里可能还残留了错误的锁定状态，这会导致用户无法抢到这个座位了！所以这里必须要做好错误处理和日志记录，确保问题可追踪，并且可以人工干预修复。
	res, err := l.svcCtx.ReleaseSeatLuaScript.Exec(l.ctx, l.svcCtx.Redis, keys, values...)
	if r, ok := res.(int64); !ok || r != 1 || err != nil {
		l.Logger.Errorf("回滚出票后，Redis 同步失败，DTM 将不断重试. order_sn: %s, err: %v, res: %v", in.OrderSn, err, res)
		return nil, status.Error(codes.Internal, "回滚出票成功，但 Redis 同步失败，请稍后重试")
	}

	return &rpc.SagaOrderResp{Success: true, Message: "出票补偿(回滚)成功"}, nil
}
