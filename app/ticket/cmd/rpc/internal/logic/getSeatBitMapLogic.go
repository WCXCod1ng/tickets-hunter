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

type GetSeatBitMapLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetSeatBitMapLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSeatBitMapLogic {
	return &GetSeatBitMapLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取BitMap信息
func (l *GetSeatBitMapLogic) GetSeatBitMap(in *rpc.GetSeatBitMapReq) (*rpc.GetSeatBitMapResp, error) {

	key := redis2.SeatBitMapRedisKey(in.EventId, in.Section)

	bitmap, err := l.svcCtx.Redis.Get(key)
	if err != nil {
		return nil, errors2.WithStack(status.Error(codes.Internal, err.Error()))
	}

	return &rpc.GetSeatBitMapResp{
		Bitmap: []byte(bitmap),
	}, nil
}
