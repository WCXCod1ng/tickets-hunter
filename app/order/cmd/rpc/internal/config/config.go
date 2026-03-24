package config

import (
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf

	// Ticket RPC 配置
	TicketRpc zrpc.RpcClientConf

	// 雪花算法配置
	Snowflake struct {
		StartTime string `json:"StartTime" binding:"required"`
		WorkerId  int64  `json:"WorkerId" binding:"required"`
	}

	// 数据库配置
	DB struct {
		DataSource string `json:"DataSource" binding:"required"`
	}

	KqPusherConf struct {
		Brokers []string `json:"Brokers" binding:"required"`
		Topic   string   `json:"Topic" binding:"required"`
	}
}
