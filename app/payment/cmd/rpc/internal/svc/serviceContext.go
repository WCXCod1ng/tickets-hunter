package svc

import (
	"tickets-hunter/app/model/user_wallet"
	"tickets-hunter/app/payment/cmd/rpc/internal/config"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config config.Config
	// 数据库连接实例
	DB sqlx.SqlConn
	// 用户钱包模型
	UserWalletModel user_wallet.UserWalletModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	mysqlConn := sqlx.NewMysql(c.DB.DataSource)

	return &ServiceContext{
		Config:          c,
		DB:              mysqlConn,
		UserWalletModel: user_wallet.NewUserWalletModel(mysqlConn),
	}
}
