package logic

import (
	"context"
	"tickets-hunter/app/ticket/model/ticket_seat"

	"tickets-hunter/app/ticket/cmd/rpc/internal/svc"
	"tickets-hunter/app/ticket/cmd/rpc/ticket/rpc"

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
	seat, err := l.svcCtx.TicketSeatModel.FindOne(l.ctx, in.SeatId)
	if err == ticket_seat.ErrNotFound {
		return nil, errors2.WithStack(status.Error(codes.NotFound, "座位不存在"))
	} else if err != nil {
		return nil, errors2.WithStack(status.Error(codes.Internal, err.Error()))
	}

	// 转换为 RPC 层的 SeatInfo
	seatInfo := &rpc.SeatInfo{
		Id:       seat.Id,
		SeatType: seat.SeatType,
		Section:  seat.Section,
		RowNo:    seat.RowNo,
		SeatNo:   seat.SeatNo,
		Price:    seat.Price,
		Status:   seat.Status,
	}

	return seatInfo, nil
}
