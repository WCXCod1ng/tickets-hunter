package svc

import (
	"tickets-hunter/app/model/order_main"
	"tickets-hunter/app/model/ticket_seat"
	"tickets-hunter/app/order/cmd/rpc/internal/config"
	"tickets-hunter/app/ticket/cmd/rpc/ticketservice"
	"tickets-hunter/common/delay_queue"
	"tickets-hunter/common/idgen"
	"tickets-hunter/common/luaexec"
	"tickets-hunter/common/mq"
	redis2 "tickets-hunter/common/redis"

	"github.com/bwmarrin/snowflake"
	"github.com/zeromicro/go-zero/core/stores/redis"
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
	// Redis
	Redis *redis.Redis
	// 延迟队列
	DelayQueue *delay_queue.ZSetDelayQueue
	// DB
	DB sqlx.SqlConn
	// 释放座位的Lua脚本
	ReleaseSeatLuaScript *luaexec.LuaScript
	// Ticket数据库模型
	TicketSeatModel ticket_seat.TicketSeatModel
	// 出座位的Lua脚本
	IssueSeatLuaScript *luaexec.LuaScript
	// MQ生产者
	MQProducer *mq.Producer
}

func NewServiceContext(c config.Config) *ServiceContext {
	mysqlConn := sqlx.NewMysql(c.DB.DataSource)

	rd := redis.MustNewRedis(c.Redis.RedisConf)

	releaseSeatScriptContent := luaexec.MustLoadLuaFile("internal/lua/releaseSeatScript.lua")
	issueSeatScriptContent := luaexec.MustLoadLuaFile("internal/lua/issueSeatScript.lua")

	return &ServiceContext{
		Config:               c,
		TicketRpc:            ticketservice.NewTicketService(zrpc.MustNewClient(c.TicketRpc)),
		Snowflake:            idgen.CreateSnowFlakeNode(c.Snowflake.StartTime, c.Snowflake.WorkerId),
		OrderMainModel:       order_main.NewOrderMainModel(mysqlConn),
		Redis:                rd,
		DelayQueue:           delay_queue.NewZSetDelayQueue(rd, redis2.OrderDelayQueueKey),
		DB:                   mysqlConn,
		ReleaseSeatLuaScript: luaexec.NewLuaScript(releaseSeatScriptContent),
		TicketSeatModel:      ticket_seat.NewTicketSeatModel(mysqlConn),
		IssueSeatLuaScript:   luaexec.NewLuaScript(issueSeatScriptContent),
		MQProducer:           mq.NewProducer(mq.NewGoZeroKafkaPusher(c.KqPusherConf.Brokers, c.KqPusherConf.Topic)),
	}
}
