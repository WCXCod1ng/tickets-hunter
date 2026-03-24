package logic

import (
	"context"
	"strconv"
	"tickets-hunter/app/model/ticket_seat"
	"tickets-hunter/app/ticket/cmd/rpc/internal/svc"
	"tickets-hunter/app/ticket/cmd/rpc/ticket/rpc"
	redis2 "tickets-hunter/common/redis"
	"time"

	errors2 "github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GetSeatInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetSeatInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSeatInfoLogic {
	return &GetSeatInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取座位信息 (供 Order RPC 内部调用)
func (l *GetSeatInfoLogic) GetSeatInfo(in *rpc.GetSeatInfoReq) (*rpc.SeatInfo, error) {
	// 优化：先查Redis，如果没有再查MySQL，同时回填Redis（Cache-Aside）
	redisKey := redis2.SeatStaticInfoRedisKey(in.SeatId)
	seatData, err := l.svcCtx.Redis.Hgetall(redisKey)
	if err == nil && len(seatData) > 0 {
		// Redis查询正常
		seatType, _ := strconv.ParseInt(seatData["seat_type"], 10, 64)
		rowNo, _ := strconv.ParseInt(seatData["row_no"], 10, 64)
		seatNo, _ := strconv.ParseInt(seatData["seat_no"], 10, 64)
		price, _ := strconv.ParseInt(seatData["price"], 10, 64)
		seatStatus, _ := strconv.ParseInt(seatData["status"], 10, 64)
		seatIndex, _ := strconv.ParseInt(seatData["seat_index"], 10, 64)
		seatInfo := &rpc.SeatInfo{
			Id:        in.SeatId,
			SeatType:  seatType,
			Section:   seatData["section"],
			RowNo:     rowNo,
			SeatNo:    seatNo,
			Price:     price,
			Status:    seatStatus,
			SeatIndex: seatIndex,
		}
		return seatInfo, nil
	}

	// 到此说明Redis查询失败，需要走MySQL
	l.Logger.Debugf("Redis查询座位信息失败，走MySQL")
	seat, err := l.svcCtx.TicketSeatModel.FindOne(l.ctx, in.SeatId)
	if err == ticket_seat.ErrNotFound {
		return nil, errors2.WithStack(status.Error(codes.NotFound, "座位不存在"))
	} else if err != nil {
		return nil, errors2.WithStack(status.Error(codes.Internal, err.Error()))
	}

	// 回写Redis
	seatMap := map[string]string{
		"event_id":   strconv.FormatInt(seat.EventId, 10),
		"seat_type":  strconv.FormatInt(seat.SeatType, 10),
		"section":    seat.Section,
		"seat_index": strconv.FormatInt(seat.SeatIndex, 10),
		"row_no":     strconv.FormatInt(seat.RowNo, 10),
		"seat_no":    strconv.FormatInt(seat.SeatNo, 10),
		"price":      strconv.FormatInt(seat.Price, 10),
	}

	err = l.svcCtx.Redis.HmsetCtx(l.ctx, redisKey, seatMap)
	if err != nil {
		l.Logger.Errorf("redis HmsetCtx err:%v", err)
	}

	// 默认缓存48小时
	expireAt := time.Now().Add(48 * time.Hour)
	if err = l.svcCtx.Redis.ExpireatCtx(l.ctx, redisKey, expireAt.Unix()); err != nil {
		l.Logger.Errorf("redis ExpireatCtx err:%v", err)
	}

	// 转换为 RPC 层的 SeatInfo
	seatInfo := &rpc.SeatInfo{
		Id:        seat.Id,
		SeatType:  seat.SeatType,
		Section:   seat.Section,
		RowNo:     seat.RowNo,
		SeatNo:    seat.SeatNo,
		Price:     seat.Price,
		Status:    seat.Status,
		SeatIndex: seat.SeatIndex,
	}

	return seatInfo, nil
}
