package svc

import (
	"tickets-hunter/app/model/usercenter"
	"tickets-hunter/app/usercenter/cmd/rpc/internal/config"
	"tickets-hunter/common/idgen"

	"github.com/bwmarrin/snowflake"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config    config.Config
	UserModel usercenter.UserModel
	Snowflake *snowflake.Node
}

func NewServiceContext(c config.Config) *ServiceContext {
	mysqlConn := sqlx.NewMysql(c.DB.DataSource)

	return &ServiceContext{
		Config:    c,
		UserModel: usercenter.NewUserModel(mysqlConn, c.UserCache),
		Snowflake: idgen.CreateSnowFlakeNode(c.Snowflake.StartTime, c.Snowflake.WorkerId),
	}
}
