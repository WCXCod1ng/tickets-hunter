package svc

import (
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"tickets-hunter/app/usercenter/cmd/rpc/internal/config"
	"tickets-hunter/app/usercenter/model"
)

type ServiceContext struct {
	Config    config.Config
	UserModel model.UserModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	mysqlConn := sqlx.NewMysql(c.DB.DataSource)

	return &ServiceContext{
		Config:    c,
		UserModel: model.NewUserModel(mysqlConn),
	}
}
