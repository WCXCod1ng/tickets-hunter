package config

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	DB struct {
		DataSource string
	}
	// TicketEventCache配置
	TicketEventCache cache.CacheConf
	// 注意不能再设置Redis配置了，因为zrpc.RpcServerConf已经包含了RPC服务的配置，RPC服务会自己去连接Redis，所以这里不需要再设置一次Redis配置了
}
