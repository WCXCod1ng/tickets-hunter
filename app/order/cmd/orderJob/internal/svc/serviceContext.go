package svc

import (
	"tickets-hunter/app/model/order_main"
	"tickets-hunter/app/order/cmd/orderJob/internal/config"
	"tickets-hunter/app/ticket/cmd/rpc/ticketservice"
	"tickets-hunter/common/delay_queue"
	redis2 "tickets-hunter/common/redis"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config config.Config
	Redis  *redis.Redis
	// DelayQueue
	DelayQueue *delay_queue.ZSetDelayQueue
	// 订单表Model
	OrderMainModel order_main.OrderMainModel
	// Ticket RPC Client
	TicketRpc ticketservice.TicketService
}

func NewServiceContext(c config.Config) *ServiceContext {

	rd := redis.MustNewRedis(c.Redis.RedisConf)

	return &ServiceContext{
		Config:         c,
		Redis:          rd,
		DelayQueue:     delay_queue.NewZSetDelayQueue(rd, redis2.OrderDelayQueueKey),
		OrderMainModel: order_main.NewOrderMainModel(sqlx.NewMysql(c.DB.DataSource)),
		TicketRpc:      ticketservice.NewTicketService(zrpc.MustNewClient(c.TicketRpc)),
	}
}
