package logic

import (
	"context"
	"tickets-hunter/app/model/ticket_seat"
	"tickets-hunter/app/ticket/define"
	redis2 "tickets-hunter/common/redis"

	"tickets-hunter/app/ticket/cmd/rpc/internal/svc"
	"tickets-hunter/app/ticket/cmd/rpc/ticket/rpc"

	errors2 "github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AcquireSeatLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAcquireSeatLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AcquireSeatLogic {
	return &AcquireSeatLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 提供完整的锁座操作
func (l *AcquireSeatLogic) AcquireSeat(in *rpc.LockSeatReq) (*rpc.LockSeatResp, error) {
	//success, err := l.svcCtx.TicketSeatModel.UpdateStatusByIdAndOldStatus(l.ctx, in.SeatId, ticket_seat.SeatStatusAvailable, ticket_seat.SeatStatusLocked)
	//if err != nil {
	//	return nil, errors2.WithStack(status.Error(codes.Internal, err.Error()))
	//}
	//return &rpc.LockSeatResp{Success: success}, nil

	// change 优化：不直接更新数据库，而是通过Redis分布式锁来控制座位的锁定状态，减少数据库压力
	bitmapKey := redis2.SeatBitMapRedisKey(in.EventId, in.Section)
	lockKey := redis2.SeatLockRedisKey(in.SeatId)
	seatIndex := in.SeatIndex
	orderSn := in.OrderSn

	// 锁单时间，单位为妙，代表支付超时时间，超过这个时间未支付则自动释放锁定的座位
	// note 注意，这个值必须超过订单超时时间，并且预留一定的时间
	lockTTL := define.SeatLockTTL // 锁定15分钟，单位为秒

	// Lua参数
	keys := []string{bitmapKey, lockKey}
	args := []any{
		seatIndex, // ARGV[1] - seatIndex
		orderSn,   // ARGV[2] - orderSn
		lockTTL,
	}

	// 执行 Lua 脚本，原子性地检查座位是否可售并锁定
	resp, err := l.svcCtx.LockSeatLuaScript.Exec(l.ctx, l.svcCtx.Redis, keys, args...)
	if err != nil {
		l.Logger.Errorf("redis lock seat err: %s", err.Error())
		return nil, errors2.WithStack(status.Error(codes.Internal, err.Error()))
	}
	res, ok := resp.(int64)
	if !ok {
		l.Logger.Errorf("unexpected redis response type: %T", resp)
		return nil, errors2.WithStack(status.Error(codes.Internal, "unexpected redis response type"))
	}

	switch res {
	case 0:
		success, err := l.svcCtx.TicketSeatModel.UpdateStatusByIdAndOldStatus(l.ctx, in.SeatId, ticket_seat.SeatStatusAvailable, ticket_seat.SeatStatusLocked)
		if err != nil || !success {
			l.Logger.Errorf("failed to update seat status in database for seat %d: %v", in.SeatId, err)
			// 尝试回滚 Redis 锁，防止锁定失效导致的座位被占用问题
			keys := []string{lockKey}
			args := []any{orderSn}
			resp, rollbackErr := l.svcCtx.UnlockSeatLuaScript.Exec(l.ctx, l.svcCtx.Redis, keys, args...)
			if r, ok := resp.(int64); !ok || r != 1 || rollbackErr != nil {
				// 回滚Redis失败了也不再重试了，此时数据库没有锁定座位，但Redis里可能还残留一个锁，这个锁会在过期时间到了之后自动失效，虽然会有短暂的锁定失效风险，但总比数据库和Redis都失效了更安全一些了
				l.Logger.Errorf("failed to rollback redis lock for seat %d: (%v, %v)", in.SeatId, resp, rollbackErr)
			}
			return nil, errors2.WithStack(status.Error(codes.Internal, "failed to lock seat"))
		}

		l.Logger.Debugf("seat %d locked successfully for order %d in Redis", in.SeatId, orderSn)
		return &rpc.LockSeatResp{Success: true}, nil
	case 1:
		// 座位不存在或不可售
		l.Logger.Debugf("seat %d is not available for locking", in.SeatId)
		return &rpc.LockSeatResp{Success: false}, nil
	case 2:
		// 座位已被锁定
		l.Logger.Debugf("seat %d is already locked", in.SeatId)
		return &rpc.LockSeatResp{Success: false}, nil
	default:
		// 其他未知错误
		l.Logger.Errorf("unexpected result from lock seat Lua script: %d", res)
		return nil, errors2.WithStack(status.Error(codes.Internal, "unexpected result from lock seat Lua script"))
	}
}
