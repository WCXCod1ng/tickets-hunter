package logic

import (
	"context"
	"tickets-hunter/app/model/ticket_seat"
	"tickets-hunter/app/ticket/cmd/rpc/internal/svc"
	"tickets-hunter/app/ticket/cmd/rpc/ticket/rpc"
	redis2 "tickets-hunter/common/redis"

	errors2 "github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ReleaseSeatLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewReleaseSeatLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReleaseSeatLogic {
	return &ReleaseSeatLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 释放座位，供订单支付超时或用户取消订单时调用，更新MySQL的座位状态为可选，并删除Redis中的锁定状态
func (l *ReleaseSeatLogic) ReleaseSeat(in *rpc.ReleaseSeatReq) (*rpc.ReleaseSeatResp, error) {
	// ==========================================
	// 1. 尝试删除 Redis 中的锁定状态
	// ==========================================
	bitmapKey := redis2.SeatBitMapRedisKey(in.EventId, in.Section)
	lockKey := redis2.SeatLockRedisKey(in.SeatId)
	seatIndex := in.SeatIndex
	orderSn := in.OrderSn
	keys := []string{bitmapKey, lockKey}
	args := []any{
		seatIndex, // ARGV[1] - seatIndex
		orderSn,   // ARGV[2] - orderSn
	}
	unlockResp, err := l.svcCtx.UnlockSeatLuaScript.Exec(
		l.ctx,
		l.svcCtx.Redis,
		keys,
		args,
	)
	if err != nil {
		// 容错处理：记录网络或 Redis 宕机异常，但【绝不 return】，继续往下走去改 MySQL！
		l.Logger.Errorf("[延迟队列] 释放 Redis 座位锁网络异常, orderSn: %s, err: %v", in.OrderSn, err)
	} else {
		// 安全类型断言
		if res, ok := unlockResp.(int64); ok {
			if res == 0 {
				// 返回 0 说明：
				// 1. 锁已经因为 15分钟 TTL 过期自动消失了；
				// 2. 发生极端并发，锁被覆盖。
				// 无论哪种情况，Redis 层面都已经没有该订单的锁了。
				l.Logger.Infof("[延迟队列] Redis 座位锁不存在或已自动过期，OrderSn: %s", in.OrderSn)
			} else {
				// 返回 1 说明主动 DEL 成功
				l.Logger.Infof("[延迟队列] Redis 座位锁主动释放成功，OrderSn: %s", in.OrderSn)
			}
		}
	}

	// ==========================================
	// 2. 兜底更新 MySQL 中的座位状态为可选
	// ==========================================
	// 使用乐观锁：UPDATE ticket_seat SET status = 0 WHERE id = ? AND status = 1
	// 无论 Redis 操作结果如何，这里都会确保数据库最终回到一致的状态。
	success, err := l.svcCtx.TicketSeatModel.UpdateStatusByIdAndOldStatus(l.ctx, in.SeatId, ticket_seat.SeatStatusLocked, ticket_seat.SeatStatusAvailable)
	if err != nil {
		l.Logger.Errorf("[延迟队列] 更新数据库座位状态失败，内部错误: %v", err)
		return nil, errors2.WithStack(status.Error(codes.Internal, err.Error()))
	}

	if !success {
		// 如果 update 影响行数为 0，说明数据库里的状态早就不是 Locked (1) 了。
		// 可能的场景：管理员后台干预、或者补偿任务已经跑过了。这是正常的幂等表现。
		l.Logger.Infof("[延迟队列] 数据库座位状态无需更新 (已被处理或不在锁定状态), SeatId: %d", in.SeatId)
	} else {
		l.Logger.Infof("[延迟队列] 数据库座位彻底释放回票池成功, SeatId: %d", in.SeatId)
	}

	return &rpc.ReleaseSeatResp{
		Success: true,
	}, nil
}
