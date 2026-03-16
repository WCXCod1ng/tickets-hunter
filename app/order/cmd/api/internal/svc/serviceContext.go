// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	"tickets-hunter/app/order/cmd/api/internal/config"
	"tickets-hunter/app/order/cmd/api/internal/middleware"
	"tickets-hunter/app/order/cmd/rpc/orderservice"

	"github.com/zeromicro/go-zero/core/limit"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config config.Config
	// OrderRpc
	OrderRpc orderservice.OrderService
	// Limiter
	Limiter *limit.TokenLimiter
	// 限流中间件
	RateLimitMiddleware rest.Middleware
}

func NewServiceContext(c config.Config) *ServiceContext {
	limiter := limit.NewTokenLimiter(
		c.TokenLimiter.Rate,
		c.TokenLimiter.Burst,
		redis.MustNewRedis(c.TokenLimiter.Redis),
		c.TokenLimiter.Key,
	)
	return &ServiceContext{
		Config:              c,
		OrderRpc:            orderservice.NewOrderService(zrpc.MustNewClient(c.OrderRpc)),
		Limiter:             limiter,
		RateLimitMiddleware: middleware.NewRateLimitMiddleware(limiter).Handle,
	}
}
