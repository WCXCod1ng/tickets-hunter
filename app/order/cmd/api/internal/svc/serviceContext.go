// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	"tickets-hunter/app/order/cmd/api/internal/config"
	"tickets-hunter/app/order/cmd/rpc/orderservice"

	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config config.Config
	// OrderRpc
	OrderRpc orderservice.OrderService
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:   c,
		OrderRpc: orderservice.NewOrderService(zrpc.MustNewClient(c.OrderRpc)),
	}
}
