package user

import (
	"context"
	"tickets-hunter/app/usercenter/cmd/rpc/usercenter/rpc"
	"tickets-hunter/common/utils"

	"tickets-hunter/app/usercenter/cmd/api/internal/svc"
	"tickets-hunter/app/usercenter/cmd/api/internal/types"

	errors2 "github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type DetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DetailLogic {
	return &DetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DetailLogic) Detail() (resp *types.UserDetailResp, err error) {
	// 1. 从上下文中获取用户ID
	userId, err := utils.GetUserIdFromToken(l.ctx)
	if err != nil {
		return nil, errors2.WithStack(status.Error(codes.Unauthenticated, "用户未登录"))
	}

	// 2. 调用RPC接口获取用户详情
	reqRpc := &rpc.DetailReq{Id: userId}
	respRpc, err := l.svcCtx.UserCenterRpc.Detail(l.ctx, reqRpc)
	if err != nil {
		return nil, errors2.WithStack(status.Error(codes.Internal, "获取用户详情失败"))
	}

	// 3. 构造响应
	resp = &types.UserDetailResp{
		Id:       respRpc.Id,
		Mobile:   respRpc.Mobile,
		Nickname: respRpc.Nickname,
		Sex:      int64(respRpc.Sex),
		Avatar:   respRpc.Avatar,
		Info:     respRpc.Info,
	}
	err = nil
	return
}
