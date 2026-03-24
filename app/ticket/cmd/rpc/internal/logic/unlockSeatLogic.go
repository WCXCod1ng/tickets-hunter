package logic

import (
	"context"
	redis2 "tickets-hunter/common/redis"

	"tickets-hunter/app/ticket/cmd/rpc/internal/svc"
	"tickets-hunter/app/ticket/cmd/rpc/ticket/rpc"

	errors2 "github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UnlockSeatLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUnlockSeatLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UnlockSeatLogic {
	return &UnlockSeatLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 取消锁定座位 (内部隐藏接口，不暴露给前端，仅供 Order RPC 调用)
func (l *UnlockSeatLogic) UnlockSeat(in *rpc.UnlockSeatReq) (*rpc.UnlockSeatResp, error) {
	bitmapKey := redis2.SeatBitMapRedisKey(in.EventId, in.Section)
	lockKey := redis2.SeatLockRedisKey(in.SeatId)
	seatIndex := in.SeatIndex
	orderSn := in.OrderSn

	// Lua参数
	keys := []string{bitmapKey, lockKey}
	args := []any{
		seatIndex, // ARGV[1] - seatIndex
		orderSn,   // ARGV[2] - orderSn
	}

	// 执行 Lua 脚本，原子性地检查订单号是否匹配并释放锁定的座位
	resp, err := l.svcCtx.UnlockSeatLuaScript.Exec(l.ctx, l.svcCtx.Redis, keys, args...)
	if err != nil {
		l.Errorf("执行 UnlockSeatLuaScript 失败: %v", err)
		return &rpc.UnlockSeatResp{Success: false}, errors2.WithStack(status.Error(codes.Internal, "释放锁定座位失败"))
	}

	res, ok := resp.(int64)
	if !ok {
		l.Errorf("UnlockSeatLuaScript 返回值类型错误: %T", resp)
		return &rpc.UnlockSeatResp{Success: false}, errors2.WithStack(status.Error(codes.Internal, "释放锁定座位失败"))
	}

	if res == 1 {
		l.Infof("成功释放锁定的座位 (seatId: %d, orderSn: %s)", in.SeatId, orderSn)
		return &rpc.UnlockSeatResp{Success: true}, nil
	}

	return &rpc.UnlockSeatResp{Success: false}, nil
}
