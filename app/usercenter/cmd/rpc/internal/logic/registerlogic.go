package logic

import (
	"context"
	"database/sql"
	"tickets-hunter/app/usercenter/model"
	"tickets-hunter/common/utils"
	"time"

	"tickets-hunter/app/usercenter/cmd/rpc/internal/svc"
	"tickets-hunter/app/usercenter/cmd/rpc/usercenter/rpc"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	if err == nil {
		return nil, status.Error(codes.AlreadyExists, "手机号已存在")
	} else if !errors.Is(err, sqlx.ErrNotFound) {
		return nil, errors.WithStack(status.Error(codes.Internal, err.Error()))
	}

	id := l.svcCtx.Snowflake.Generate().Int64()

	password, err := utils.Encrypt(in.GetPassword())
	if err != nil {
		return nil, errors.WithStack(status.Error(codes.Internal, err.Error()))
	}

	// 到此说明手机号不重复
	// 生成并插入数据
	user := &model.User{
		Id:         id,
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
		DeleteTime: sql.NullTime{},
		Mobile:     in.GetMobile(),
		Password:   password,
		Nickname:   sql.NullString{},
		Sex:        0,
		Avatar:     sql.NullString{},
		Info:       sql.NullString{},
	}
	if _, err := l.svcCtx.UserModel.Insert(l.ctx, user); err != nil {
		return nil, errors.WithStack(status.Error(codes.Internal, err.Error()))
	}

	return &rpc.RegisterResp{
		Id:          user.Id,
		Token:       "token",
		TokenExpire: 0,
	}, nil
}
