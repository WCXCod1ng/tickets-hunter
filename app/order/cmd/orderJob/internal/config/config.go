package config

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
	zrpc.RpcServerConf
	DB struct {
		DataSource string
	}

	// Ticket RPC Client 配置
	TicketRpc zrpc.RpcClientConf
}
