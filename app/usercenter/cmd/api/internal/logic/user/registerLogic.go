package user

import (
	"context"
	"tickets-hunter/app/usercenter/cmd/rpc/usercenter/rpc"

	"tickets-hunter/app/usercenter/cmd/api/internal/svc"
	"tickets-hunter/app/usercenter/cmd/api/internal/types"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type RegisterLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RegisterLogic) Register(req *types.RegisterReq) (resp *types.LoginRegisterResp, err error) {
	reqRpc := &rpc.RegisterReq{
		Mobile:   req.Mobile,
		Password: req.Password,
		Nickname: req.Nickname,
		Sex:      req.Sex,
		Avatar:   req.Avatar,
		Info:     req.Info,
	}
	respRpc, err := l.svcCtx.UserCenterRpc.Register(l.ctx, reqRpc)
	if err != nil {
		// 对于Rpc端的错误，我们应当通过Wrapf来包装
		return nil, errors.WithStack(err)
	}

	resp = &types.LoginRegisterResp{}

	resp.Id = respRpc.Id
	resp.Token = respRpc.Token
	resp.TokenExpire = respRpc.TokenExpire
	err = nil
	return
}
