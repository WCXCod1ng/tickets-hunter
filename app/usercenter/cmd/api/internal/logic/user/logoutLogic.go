// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"context"
	"fmt"
	"tickets-hunter/app/usercenter/cmd/api/internal/entity"

	"tickets-hunter/app/usercenter/cmd/api/internal/svc"
	"tickets-hunter/app/usercenter/cmd/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type LogoutLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 用户登出
func NewLogoutLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LogoutLogic {
	return &LogoutLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LogoutLogic) Logout(req *types.LogoutReq) error {
	// 1. 先查一下这个 Token 属于哪个用户 (为了去集合里删除它)
	refreshKey := fmt.Sprintf("usercenter:auth:refresh:%s", req.RefreshToken)
	refreshValue, err := l.svcCtx.Redis.Get(refreshKey)

	// 如果 Token 已经过期或不存在，直接返回成功即可
	if err != nil || refreshValue == "" {
		return nil
	}

	// 解析出用户ID
	var refreshInfo entity.RefreshInfo
	if err := l.svcCtx.Serializer.Unmarshal([]byte(refreshValue), &refreshInfo); err != nil {
		return nil // 解析失败也当作成功处理，反正这个 Token 已经失效了
	}
	userIdStr := refreshInfo.UserId

	// 2. 删除 Token 本身 (让它失效)
	l.svcCtx.Redis.Del(refreshKey)

	// 3. 从用户的 Token 集合中移除这个 Token (清理索引)
	userTokenKey := fmt.Sprintf("usercenter:auth:user:refresh_tokens:%s", userIdStr)
	l.svcCtx.Redis.Srem(userTokenKey, req.RefreshToken)

	return nil
}
