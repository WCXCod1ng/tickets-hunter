package logic

import (
	"context"

	"tickets-hunter/app/ticket/cmd/rpc/internal/svc"
	"tickets-hunter/app/ticket/cmd/rpc/ticket/rpc"

	errors2 "github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GetSeatListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetSeatListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSeatListLogic {
	return &GetSeatListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取指定场次的座位图及状态 (供 API 层调用)
func (l *GetSeatListLogic) GetSeatList(in *rpc.GetSeatListReq) (*rpc.GetSeatListResp, error) {
	ticketSeats, err := l.svcCtx.TicketSeatModel.FindByEventId(l.ctx, in.EventId)
	if err != nil {
		return nil, errors2.WithStack(status.Error(codes.Internal, err.Error()))
	}

	list := make([]*rpc.SeatInfo, len(ticketSeats))
	for i, seat := range ticketSeats {
		list[i] = &rpc.SeatInfo{
			Id:       seat.Id,
			SeatType: seat.SeatType,
			Section:  seat.Section,
			RowNo:    seat.RowNo,
			SeatNo:   seat.SeatNo,
			Price:    seat.Price,
			Status:   seat.Status,
		}
	}

	return &rpc.GetSeatListResp{
		List: list,
	}, nil
}
