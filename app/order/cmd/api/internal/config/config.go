// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package config

import (
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	rest.RestConf
	Auth struct {
		AccessSecret string
		AccessExpire int64
	}
	// OrderRpc配置
	OrderRpc zrpc.RpcClientConf `json:"OrderRpc"`
	// Redis配置
	Redis redis.RedisConf `json:"Redis"`
	//// CacheRedis配置
	//CacheRedis redis.RedisConf `json:"CacheRedis"`
	// 令牌桶限流器配置
	TokenLimiter struct {
		Redis   redis.RedisConf
		Key     string // Redis中存储令牌桶状态的键
		Rate    int    // 每秒生成的令牌数
		Burst   int    // 令牌桶的容量
		Seconds int    // 时间窗口大小，单位为秒
	}
}
