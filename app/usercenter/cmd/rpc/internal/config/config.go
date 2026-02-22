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
	Snowflake struct {
		StartTime string `json:"StartTime" binding:"required"`
		WorkerId  int64  `json:"WorkerId" binding:"required"`
	}
	// Cache配置
	UserCache cache.CacheConf
}
