package ticket_event

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TicketEventModel = (*customTicketEventModel)(nil)

type (
	// TicketEventModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTicketEventModel.
	TicketEventModel interface {
		ticketEventModel
		withSession(session sqlx.Session) TicketEventModel
		// 根据status查询场次列表
		FindByStatus(ctx context.Context, status int64) ([]*TicketEvent, error)
	}

	customTicketEventModel struct {
		*defaultTicketEventModel
	}
)

func (m *customTicketEventModel) FindByStatus(ctx context.Context, status int64) ([]*TicketEvent, error) {
	query := fmt.Sprintf("select %s from %s where status = ?", ticketEventRows, m.table)
	var resp []*TicketEvent
	err := m.conn.QueryRowsCtx(ctx, &resp, query, status)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// NewTicketEventModel returns a usercenter for the database table.
func NewTicketEventModel(conn sqlx.SqlConn) TicketEventModel {
	return &customTicketEventModel{
		defaultTicketEventModel: newTicketEventModel(conn),
	}
}

func (m *customTicketEventModel) withSession(session sqlx.Session) TicketEventModel {
	return NewTicketEventModel(sqlx.NewSqlConnFromSession(session))
}
