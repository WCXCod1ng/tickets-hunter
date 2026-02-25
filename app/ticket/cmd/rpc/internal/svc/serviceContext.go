package svc

import (
	"tickets-hunter/app/ticket/cmd/rpc/internal/config"
	"tickets-hunter/app/ticket/model/ticket_event"
	"tickets-hunter/app/ticket/model/ticket_seat"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config           config.Config
	TicketEventModel ticket_event.TicketEventModel
	TicketSeatModel  ticket_seat.TicketSeatModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	mysqlConn := sqlx.NewMysql(c.DB.DataSource)

	return &ServiceContext{
		Config:           c,
		TicketEventModel: ticket_event.NewTicketEventModel(mysqlConn),
		TicketSeatModel:  ticket_seat.NewTicketSeatModel(mysqlConn),
	}
}
