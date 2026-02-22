package config

import (
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	rest.RestConf
	Auth struct {
		AccessSecret  string
		AccessExpire  int64
		RefreshExpire int64
	}
	// 如果需要依赖RPC，那么通过这种方式导入RPC的配置
	UserCenterRpc zrpc.RpcClientConf `json:"UserCenterRpc"`
	Redis         redis.RedisConf    `json:"Redis"`
}
