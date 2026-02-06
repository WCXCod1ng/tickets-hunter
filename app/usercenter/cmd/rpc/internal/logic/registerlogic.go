package logic

import (
	"context"
	"database/sql"
	"tickets-hunter/app/usercenter/model"
	"time"

	"tickets-hunter/app/usercenter/cmd/rpc/internal/svc"
	"tickets-hunter/app/usercenter/cmd/rpc/usercenter/rpc"

	"github.com/zeromicro/go-zero/core/logx"
)

type RegisterLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RegisterLogic) Register(in *rpc.RegisterReq) (*rpc.RegisterResp, error) {
	// 校验手机号是否已经存在
	_, err := l.svcCtx.UserModel.FindOneByMobile(l.ctx, in.GetMobile())
	if err != nil {
		return nil, err
	}
	// 生成并插入数据
	user := &model.User{
		Id:         0,
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
		DeleteTime: sql.NullTime{},
		Mobile:     in.GetMobile(),
		Password:   in.GetPassword(),
		Nickname:   sql.NullString{},
		Sex:        0,
		Avatar:     sql.NullString{},
		Info:       sql.NullString{},
	}
	if _, err := l.svcCtx.UserModel.Insert(l.ctx, user); err != nil {
		return nil, err
	}

	return &rpc.RegisterResp{
		Id:          user.Id,
		Token:       "token",
		TokenExpire: 0,
	}, nil
}
