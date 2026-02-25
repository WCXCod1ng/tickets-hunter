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

type GetEventListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetEventListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetEventListLogic {
	return &GetEventListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取演唱会场次列表 (供 API 层调用)
func (l *GetEventListLogic) GetEventList(in *rpc.GetEventListReq) (*rpc.GetEventListResp, error) {
	ticketEvents, err := l.svcCtx.TicketEventModel.FindByStatus(l.ctx, 1)
	if err != nil {
		return nil, errors2.WithStack(status.Error(codes.Internal, err.Error()))
	}

	list := make([]*rpc.EventInfo, 0, len(ticketEvents))
	for _, event := range ticketEvents {
		list = append(list, &rpc.EventInfo{
			Id:            event.Id,
			Title:         event.Title,
			CoverUrl:      event.CoverUrl,
			ShowTime:      event.ShowTime.String(),
			Venue:         event.Venue,
			SaleStartTime: event.SaleStartTime.String(),
			SaleEndTime:   event.SaleEndTime.String(),
			Status:        event.Status,
		})
	}

	return &rpc.GetEventListResp{
		List:  list,
		Total: int64(len(list)),
	}, nil
}
