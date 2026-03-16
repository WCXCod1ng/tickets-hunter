package logic

import (
	"context"
	"errors"
	"tickets-hunter/app/model/ticket_seat"
	redis2 "tickets-hunter/common/redis"
	"time"

	"tickets-hunter/app/ticket/cmd/rpc/internal/svc"
	"tickets-hunter/app/ticket/cmd/rpc/ticket/rpc"

	errors2 "github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type WarmUpValidSeatsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewWarmUpValidSeatsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *WarmUpValidSeatsLogic {
	return &WarmUpValidSeatsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 有效座位集合缓存预热 (内部隐藏接口，实际场景中可能由定时任务调用)
func (l *WarmUpValidSeatsLogic) WarmUpValidSeats(in *rpc.WarmUpValidSeatsReq) (*rpc.WarmUpValidSeatsResp, error) {
	// 1. 查询所有有效座位（状态为可售的座位）
	// 查询演唱会场次表，获取演唱会场次信息
	ticketEvent, err := l.svcCtx.TicketEventModel.FindOne(l.ctx, in.EventId)
	if errors.Is(err, ticket_seat.ErrNotFound) {
		return nil, status.Error(codes.NotFound, "演唱会场次不存在")
	} else if err != nil {
		return nil, errors2.WithStack(status.Error(codes.Internal, err.Error()))
	}

	// 前置条件检查：如果演唱会场次已结束，则不进行缓存预热
	if time.Now().After(ticketEvent.SaleEndTime) {
		return nil, status.Error(codes.InvalidArgument, "演唱会场次已结束，不需要预热缓存")
	}

	seats, err := l.svcCtx.TicketSeatModel.FindByEventId(l.ctx, in.EventId)
	if err != nil {
		return nil, errors2.WithStack(status.Error(codes.Internal, err.Error()))
	}
	if len(seats) == 0 {
		return nil, status.Error(codes.NotFound, "没有有效座位")
	}

	// 2. 批量将有效座位信息写入 Redis 缓存

	// 3. 按 Section 分组
	sectionMap := make(map[string][]int64)
	for _, seat := range seats {
		sectionMap[seat.Section] = append(sectionMap[seat.Section], seat.SeatIndex)
	}

	// 4. 对每个 Section 写入 Bitmap
	for section, seatIndexes := range sectionMap {
		redisKey := redis2.SeatBitMapRedisKey(in.EventId, section)

		// 删除旧缓存
		if _, err := l.svcCtx.Redis.Del(redisKey); err != nil {
			return nil, errors2.WithStack(err)
		}

		totalCached := 0
		batchSize := 1000
		if err := l.svcCtx.Redis.PipelinedCtx(l.ctx, func(pipe redis.Pipeliner) error {
			for i := 0; i < len(seatIndexes); i += batchSize {
				end := i + batchSize
				if end > len(seatIndexes) {
					end = len(seatIndexes)
				}
				for _, seatIndex := range seatIndexes[i:end] {
					pipe.SetBit(l.ctx, redisKey, seatIndex, 0) // BitMap中：0-可售；1-不可售（卖出/锁定）
				}
				totalCached += end - i
			}
			return nil
		}); err != nil {
			return nil, errors2.WithStack(err)
		}

		// 设置过期时间
		expireAt := ticketEvent.SaleEndTime.Add(3 * time.Hour)
		if err := l.svcCtx.Redis.ExpireatCtx(l.ctx, redisKey, expireAt.Unix()); err != nil {
			return nil, errors2.WithStack(err)
		}

		l.Logger.Infof("预热 Section %s Bitmap 成功，写入 %d 个座位", section, totalCached)
	}

	return &rpc.WarmUpValidSeatsResp{Success: true}, nil
}
