// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	"tickets-hunter/app/model/order_main"
	"tickets-hunter/app/payment/cmd/api/internal/config"
	"tickets-hunter/app/payment/cmd/rpc/payment"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config         config.Config
	PaymentRpc     payment.Payment
	OrderModelMain order_main.OrderMainModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	mysqlconn := sqlx.NewMysql(c.DB.DataSource)

	return &ServiceContext{
		Config:         c,
		PaymentRpc:     payment.NewPayment(zrpc.MustNewClient(c.PaymentRpc)),
		OrderModelMain: order_main.NewOrderMainModel(mysqlconn),
	}
}
