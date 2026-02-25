// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package ticket

import (
	"context"
	"tickets-hunter/app/ticket/cmd/rpc/ticket/rpc"

	"tickets-hunter/app/ticket/cmd/api/internal/svc"
	"tickets-hunter/app/ticket/cmd/api/internal/types"

	errors2 "github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GetEventListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取上架的演唱会场次列表
func NewGetEventListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetEventListLogic {
	return &GetEventListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetEventListLogic) GetEventList() (resp *types.EventListResp, err error) {
	// 调用 RPC 层获取数据
	rpcResp, err := l.svcCtx.TicketRpc.GetEventList(l.ctx, &rpc.GetEventListReq{})
	if err != nil {
		return nil, errors2.WithStack(status.Error(codes.Internal, err.Error()))
	}

	// 转换 RPC 层的响应为 API 层的响应
	eventList := make([]types.EventInfo, 0, len(rpcResp.List))
	for _, event := range rpcResp.List {
		eventList = append(eventList, types.EventInfo{
			Id:            event.Id,
			Title:         event.Title,
			CoverUrl:      event.CoverUrl,
			ShowTime:      event.ShowTime,
			Venue:         event.Venue,
			SaleStartTime: event.SaleStartTime,
			SaleEndTime:   event.SaleEndTime,
			Status:        event.Status,
		})
	}

	return &types.EventListResp{
		List:  eventList,
		Total: rpcResp.Total,
	}, nil
}
