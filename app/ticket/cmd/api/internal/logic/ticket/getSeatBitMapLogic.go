// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package ticket

import (
	"context"
	"encoding/base64"
	"tickets-hunter/app/ticket/cmd/rpc/ticketservice"
	redis2 "tickets-hunter/common/redis"

	"tickets-hunter/app/ticket/cmd/api/internal/svc"
	"tickets-hunter/app/ticket/cmd/api/internal/types"

	errors2 "github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GetSeatBitMapLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取指定场次的座位图（BitMap形式）及状态
func NewGetSeatBitMapLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSeatBitMapLogic {
	return &GetSeatBitMapLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetSeatBitMapLogic) GetSeatBitMap(req *types.SeatBitMapReq) (resp *types.SeatBitMapResp, err error) {
	rpcResp, err := l.svcCtx.TicketRpc.GetSeatBitMap(l.ctx, &ticketservice.GetSeatBitMapReq{
		EventId: req.EventId,
		Section: req.Section,
	})
	if err != nil {
		return nil, errors2.WithStack(status.Error(codes.Internal, err.Error()))
	}

	// Redis中查Layout信息
	layout, err := l.svcCtx.Redis.Get(redis2.SectionSeatStaticInfoRedisKey(req.EventId, req.Section))
	if err != nil {
		return nil, errors2.WithStack(status.Error(codes.Internal, err.Error()))
	}

	return &types.SeatBitMapResp{
		Layout: layout,
		BitMap: base64.StdEncoding.EncodeToString(rpcResp.Bitmap),
	}, nil
}
