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

type IssueTicketLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewIssueTicketLogic(ctx context.Context, svcCtx *svc.ServiceContext) *IssueTicketLogic {
	return &IssueTicketLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Saga 步骤2 正向操作：确认出票
func (l *IssueTicketLogic) IssueTicket(in *rpc.SagaOrderReq) (*rpc.SagaOrderResp, error) {
	l.Logger.Debugf("开始执行出票正向操作，orderSn: %s", in.OrderSn)
	barrier, err := dtmgrpc.BarrierFromGrpc(l.ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	db, err := l.svcCtx.DB.RawDB()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// 针对DTM空补偿的问题，我们需要在这里区分两种错误类型：
	// 业务异常：比如订单状态不对、座位状态不对等，这些都是我们预期内可能发生的错误，应该直接返回给调用方，让 DTM 触发补偿逻辑（回滚出票），并且这些错误不应该被重试，因为重试也无法成功。
	// 系统异常：比如数据库连接失败、SQL 执行错误等，这些是我们预期外的错误，可能是暂时性的，应该记录日志并返回系统异常错误，让 DTM 重试这个步骤，直到成功或者达到重试上限。
	// 🌟 核心修改 1：定义一个外部变量，用于捕获“业务异常”
	var bizErr error

	err = barrier.CallWithDB(db, func(tx *sql.Tx) error {
		// 1. 推进订单状态：10(待支付) -> 30(已出票)
		success, err := l.svcCtx.OrderMainModel.UpdateStatusByOrderSnAndStatusWithTx(l.ctx, tx, in.OrderSn, order_main.OrderStatusPending, order_main.OrderStatusIssued)
		if err != nil {
			return err // 系统异常，让 DTM 重试
		}

		if !success {
			// 🌟 核心修改 2：只记录业务异常，但向 CallWithDB 返回 nil！
			// 返回 nil 会让本地事务 COMMIT，成功把 action 屏障落库！
			l.Logger.Errorf("出票失败：订单状态异常, order_sn: %s", in.OrderSn)
			bizErr = status.Error(codes.Aborted, "订单状态异常，自动触发退款")
			return nil
		}

		// 2. 推进座位状态：1(锁定) -> 2(已售)
		success, err = l.svcCtx.TicketSeatModel.UpdateStatusByIdAndOldStatusWithTx(l.ctx, tx, in.SeatId, ticket_seat.SeatStatusLocked, ticket_seat.SeatStatusSold)
		if err != nil {
			return err
		}

		if !success {
			l.Logger.Errorf("出票失败：座位状态异常, seat_id: %d", in.SeatId)
			bizErr = status.Error(codes.Aborted, "座位状态异常，自动触发退款")
			return nil // 同理，返回 nil 提交屏障
		}

		return nil // 订单和座位同时更新成功！
	})

	// 如果是系统异常，直接返回错误让 DTM 重试；如果是业务异常，返回特定错误让 DTM 触发补偿逻辑（回滚出票）
	if err != nil {
		return &rpc.SagaOrderResp{Success: false, Message: "系统异常，DTM 将重试"}, status.Error(codes.Internal, err.Error())
	}

	if bizErr != nil {
		return &rpc.SagaOrderResp{Success: false, Message: "业务异常，DTM 将触发补偿"}, bizErr
	}

	// 出票成功，从Redis中删除。注意这一步无论成功与否都不能让DTM回退：如果你返回了 error，DTM 会认为出票失败，直接去执行退款！这就变成了**“票已经出在库里了，钱却退给用户了”**的严重资损
	// 而且使用异步 Goroutine 来执行这个清理操作，绝对不能阻塞主流程的返回，否则会增加 DTM 的整体响应时间，降低系统性能和用户体验！
	go func() {
		// 重新生成 Context，因为外部的 ctx 可能马上会被 cancel
		bgCtx := context.Background()

		bitmapKey := redis2.SeatBitMapRedisKey(in.EventId, in.Section)
		lockKey := redis2.SeatLockRedisKey(in.SeatId)
		l.Logger.Infof("bitmapKey: %s, lockKey: %s", bitmapKey, lockKey)

		keys := []string{
			bitmapKey,
			lockKey,
			redis2.OrderDelayQueueKey,
		}
		args := []any{
			in.SeatIndex, // ARGV[1] - seatIndex
			in.OrderSn,   // ARGV[2] - orderSn
		}
		// 执行 Lua 脚本，原子性地清理 Redis 中的座位锁定、有效性状态、以及订单延时队列中的相关消息
		res, err := l.svcCtx.IssueSeatLuaScript.Exec(bgCtx, l.svcCtx.Redis, keys, args...)

		if r, ok := res.(int64); !ok || r != 1 || err != nil {
			// ⚠️ 仅仅打印告警日志，绝对不影响主流程！
			l.Logger.Errorf("出票成功但清理 Redis 失败, 需要人工/定时任务介入清理脏缓存. seatId: %d orderSn: %s, err: %v, res: %v", in.SeatId, in.OrderSn, err, res)
		} else {
			l.Logger.Infof("座位 %d 的 Redis 锁定和有效性状态已成功清理, orderSn: %s", in.SeatId, in.OrderSn)
		}
	}()

	return &rpc.SagaOrderResp{Success: true, Message: "出票成功"}, nil
}
