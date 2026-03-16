package logic

import (
	"context"
	"errors"
	"tickets-hunter/app/model/usercenter"
	"tickets-hunter/app/usercenter/cmd/rpc/internal/svc"
	"tickets-hunter/app/usercenter/cmd/rpc/usercenter/rpc"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type DetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DetailLogic {
	return &DetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DetailLogic) Detail(in *rpc.DetailReq) (*rpc.DetailResp, error) {
	user, err := l.svcCtx.UserModel.FindOne(l.ctx, in.Id)
	if err != nil {
		if errors.Is(err, usercenter.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "用户不存在")
		}
		return nil, err
	}

	return &rpc.DetailResp{
		Id:       user.Id,
		Mobile:   user.Mobile,
		Nickname: user.Nickname.String,
		Sex:      int32(user.Sex),
		Avatar:   user.Avatar.String,
		Info:     user.Info.String,
	}, nil
}
