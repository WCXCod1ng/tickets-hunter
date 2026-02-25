// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	"tickets-hunter/app/ticket/cmd/api/internal/config"
	"tickets-hunter/app/ticket/cmd/rpc/ticketservice"

	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config    config.Config
	TicketRpc ticketservice.TicketService
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:    c,
		TicketRpc: ticketservice.NewTicketService(zrpc.MustNewClient(c.TicketRpc)),
	}
}
