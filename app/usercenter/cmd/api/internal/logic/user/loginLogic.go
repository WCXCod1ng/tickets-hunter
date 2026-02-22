package user

import (
	"context"
	"fmt"
	"strconv"
	"tickets-hunter/app/usercenter/cmd/api/internal/entity"
	"tickets-hunter/app/usercenter/cmd/rpc/usercenter/rpc"
	"tickets-hunter/common/utils"
	"time"

	"tickets-hunter/app/usercenter/cmd/api/internal/svc"
	"tickets-hunter/app/usercenter/cmd/api/internal/types"

	"github.com/google/uuid"
	errors2 "github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type LoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginLogic) Login(req *types.LoginReq) (resp *types.LoginRegisterResp, err error) {
	reqRpc := &rpc.LoginReq{
		Mobile:   req.Mobile,
		Password: req.Password,
	}
	respRpc, err := l.svcCtx.UserCenterRpc.Login(l.ctx, reqRpc)
	if err != nil {
		return nil, err
	}

	// 登录成功

	now := time.Now().Unix()
	// 生成Access Token
	token, err := utils.GenerateJwtToken(l.svcCtx.Config.Auth.AccessSecret, now, l.svcCtx.Config.Auth.AccessExpire, fmt.Sprintf("%d", respRpc.Id))
	if err != nil {
		return nil, errors2.WithStack(status.Error(codes.Internal, err.Error()))
	}

	// 生成Refresh Token并存入Redis
	refreshToken := uuid.New().String()
	refreshExpire := l.svcCtx.Config.Auth.RefreshExpire
	refreshInfo := entity.RefreshInfo{
		UserId:   strconv.FormatInt(respRpc.Id, 10),
		Platform: "web", // TODO 这里可以根据实际情况设置平台类型
	}
	var value []byte
	if value, err = l.svcCtx.Serializer.Marshal(refreshInfo); err != nil {
		return nil, errors2.WithStack(status.Error(codes.Internal, err.Error()))
	}
	verifyKey := fmt.Sprintf("usercenter:auth:refresh:%s", refreshToken)
	if err := l.svcCtx.Redis.Setex(verifyKey, string(value), int(refreshExpire)); err != nil {
		return nil, errors2.WithStack(status.Error(codes.Internal, err.Error()))
	}

	// 同时将Refresh Token的Key存入Redis，方便后续验证和管理
	controlKey := fmt.Sprintf("usercenter:auth:user:refresh_tokens:%d", respRpc.Id)
	if _, err := l.svcCtx.Redis.Sadd(controlKey, refreshToken); err != nil {
		return nil, errors2.WithStack(status.Error(codes.Internal, err.Error()))
	}
	if err := l.svcCtx.Redis.Expire(controlKey, int(refreshExpire)); err != nil {
		return nil, errors2.WithStack(status.Error(codes.Internal, err.Error()))
	}

	resp = &types.LoginRegisterResp{
		Token:        token,
		TokenExpire:  now + l.svcCtx.Config.Auth.AccessExpire,
		RefreshToken: refreshToken,
	}

	return
}
