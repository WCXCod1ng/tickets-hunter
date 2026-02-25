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

type GetSeatListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取指定场次的座位图及状态
func NewGetSeatListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSeatListLogic {
	return &GetSeatListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetSeatListLogic) GetSeatList(req *types.SeatListReq) (resp *types.SeatListResp, err error) {
	// 调用 RPC 层获取数据
	rpcResp, err := l.svcCtx.TicketRpc.GetSeatList(l.ctx, &rpc.GetSeatListReq{
		EventId: req.EventId,
	})
	if err != nil {
		return nil, errors2.WithStack(status.Error(codes.Internal, err.Error()))
	}

	// 转换 RPC 层的响应为 API 层的响应
	seatList := make([]types.SeatInfo, 0, len(rpcResp.List))
	for _, seat := range rpcResp.List {
		seatList = append(seatList, types.SeatInfo{
			Id:       seat.Id,
			SeatType: seat.SeatType,
			Section:  seat.Section,
			RowNo:    seat.RowNo,
			SeatNo:   seat.SeatNo,
			Price:    float64(seat.Price) / 100, // 转换为元
			Status:   seat.Status,
		})
	}

	return &types.SeatListResp{
		List: seatList,
	}, nil
}
