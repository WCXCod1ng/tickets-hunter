// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"tickets-hunter/app/usercenter/cmd/api/internal/entity"
	"tickets-hunter/common/utils"
	"time"

	"tickets-hunter/app/usercenter/cmd/api/internal/svc"
	"tickets-hunter/app/usercenter/cmd/api/internal/types"

	errors2 "github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type RefreshLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// token刷新
func NewRefreshLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RefreshLogic {
	return &RefreshLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RefreshLogic) Refresh(req *types.RefreshReq) (resp *types.RefreshResp, err error) {
	// 根据UUID从Redis中获取Refresh Token信息
	verifyKey := fmt.Sprintf("usercenter:auth:refresh:%s", req.RefreshToken)
	value, err := l.svcCtx.Redis.Get(verifyKey)
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, errors2.WithStack(status.Error(codes.Internal, err.Error()))
	}

	// 如果Redis中没有对应的Refresh Token，说明无效或已过期
	if err != nil || value == "" {
		return nil, status.Error(codes.Unauthenticated, "Refresh Token无效或已过期，请重新登录")
	}

	// 反之说明Refresh Token有效，解析出其中的信息
	var refreshInfo entity.RefreshInfo
	if err := l.svcCtx.Serializer.Unmarshal([]byte(value), &refreshInfo); err != nil {
		return nil, errors2.WithStack(status.Error(codes.Internal, err.Error()))
	}
	userId, err := strconv.ParseInt(refreshInfo.UserId, 10, 64)
	if err != nil {
		return nil, errors2.WithStack(status.Error(codes.Internal, err.Error()))
	}

	// 生成新的Access Token
	now := time.Now().Unix()
	accessToken, err := utils.GenerateJwtToken(l.svcCtx.Config.Auth.AccessSecret, now, l.svcCtx.Config.Auth.AccessExpire, fmt.Sprintf("%d", userId))
	if err != nil {
		return nil, errors2.WithStack(status.Error(codes.Internal, err.Error()))
	}

	// TODO 这里可以根据实际需求决定是否要更新Refresh Token的过期时间或者生成新的Refresh Token，目前的实现是继续使用原来的Refresh Token，直到它过期为止

	resp = &types.RefreshResp{
		AccessToken:  accessToken,
		AccessExpire: now + l.svcCtx.Config.Auth.AccessExpire,
		RefreshToken: req.RefreshToken, // 刷新后Refresh Token不变，继续使用原来的
	}

	return
}
