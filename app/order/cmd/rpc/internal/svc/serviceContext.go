package svc

import (
	"tickets-hunter/app/order/cmd/rpc/internal/config"
	"tickets-hunter/app/order/model/order_main"
	"tickets-hunter/app/ticket/cmd/rpc/ticketservice"
	"tickets-hunter/common/idgen"

	"github.com/bwmarrin/snowflake"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config config.Config
	// 用于调用 Ticket RPC 服务，获取座位信息和锁定座位
	TicketRpc ticketservice.TicketService
	// 雪花算法
	Snowflake *snowflake.Node
	// OrderMainModel
	OrderMainModel order_main.OrderMainModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	mysqlConn := sqlx.NewMysql(c.DB.DataSource)

	return &ServiceContext{
		Config:         c,
		TicketRpc:      ticketservice.NewTicketService(zrpc.MustNewClient(c.TicketRpc)),
		Snowflake:      idgen.CreateSnowFlakeNode(c.Snowflake.StartTime, c.Snowflake.WorkerId),
		OrderMainModel: order_main.NewOrderMainModel(mysqlConn),
	}
}
