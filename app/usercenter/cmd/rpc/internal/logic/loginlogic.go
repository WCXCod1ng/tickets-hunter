package logic

import (
	"context"
	"errors"
	"tickets-hunter/app/usercenter/model"
	"tickets-hunter/common/utils"
	"tickets-hunter/common/xerr"

	"tickets-hunter/app/usercenter/cmd/rpc/internal/svc"
	"tickets-hunter/app/usercenter/cmd/rpc/usercenter/rpc"

	errors2 "github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type LoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *LoginLogic) Login(in *rpc.LoginReq) (*rpc.LoginResp, error) {
	user, err := l.svcCtx.UserModel.FindOneByMobile(l.ctx, in.Mobile)
	if errors.Is(err, model.ErrNotFound) {
		return nil, status.Error(codes.NotFound, "用户不存在")
	} else if err != nil {
		return nil, errors2.WithStack(status.Error(codes.Internal, err.Error()))
	}

	// 验证密码
	if !utils.Verify(in.Password, user.Password) {
		return nil, status.Error(codes.Code(xerr.LoginFailed), "账号或密码错误")
	}

	return &rpc.LoginResp{
		Id: user.Id,
	}, nil
}
