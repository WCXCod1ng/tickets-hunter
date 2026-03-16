package ticket_seat

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TicketSeatModel = (*customTicketSeatModel)(nil)

type (
	// TicketSeatModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTicketSeatModel.
	TicketSeatModel interface {
		ticketSeatModel
		withSession(session sqlx.Session) TicketSeatModel

		// 通过ticketEventId查询座位信息
		FindByEventId(ctx context.Context, ticketEventId int64) ([]*TicketSeat, error)

		// 根据Id和作为状态更新座位状态
		UpdateStatusByIdAndOldStatus(ctx context.Context, id int64, status int64, newStatus int64) (bool, error)

		// 带事务的根据Id和作为状态更新座位状态，主要用于订单支付成功后更新座位状态的场景
		UpdateStatusByIdAndOldStatusWithTx(ctx context.Context, tx *sql.Tx, id int64, status int64, newStatus int64) (bool, error)

		UpdateStatusByIdAndInStatusWithTx(ctx context.Context, tx *sql.Tx, id int64, inStatus []int64, newStatus int64) (bool, error)

		// 根据演唱会id查询座位id列表（仅限未售出的座位）
		FindValidIdsByEventId(ctx context.Context, eventId int64) ([]int64, error)
	}

	customTicketSeatModel struct {
		*defaultTicketSeatModel
	}
)

func (m *customTicketSeatModel) FindByEventId(ctx context.Context, ticketEventId int64) ([]*TicketSeat, error) {
	query := fmt.Sprintf("select %s FROM %s WHERE event_id = ?", ticketSeatRows, m.table)
	var resp []*TicketSeat
	err := m.conn.QueryRowsCtx(ctx, &resp, query, ticketEventId)
	if err != nil {
		return nil, err
	}
	return resp, err
}

func (m *customTicketSeatModel) UpdateStatusByIdAndOldStatus(ctx context.Context, id int64, oldStatus int64, newStatus int64) (bool, error) {
	query := fmt.Sprintf("update %s set status = ? where id = ? and status = ?", m.table)
	result, err := m.conn.ExecCtx(ctx, query, newStatus, id, oldStatus)
	if err != nil {
		return false, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	return rowsAffected > 0, nil
}

func (m *customTicketSeatModel) UpdateStatusByIdAndOldStatusWithTx(ctx context.Context, tx *sql.Tx, id int64, status int64, newStatus int64) (bool, error) {
	query := fmt.Sprintf("update %s set status = ? where id = ? and status = ?", m.table)
	result, err := tx.ExecContext(ctx, query, newStatus, id, status)
	if err != nil {
		return false, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	return rowsAffected > 0, nil
}

func (m *customTicketSeatModel) UpdateStatusByIdAndInStatusWithTx(ctx context.Context, tx *sql.Tx, id int64, inStatus []int64, newStatus int64) (bool, error) {
	// 构建 IN 条件的占位符
	placeholders := make([]string, len(inStatus))
	args := make([]interface{}, len(inStatus)+2) // +2 用于 id 和 newStatus
	args[0] = newStatus
	args[1] = id
	for i, status := range inStatus {
		placeholders[i] = "?"
		args[i+2] = status
	}

	query := fmt.Sprintf("update %s set status = ? where id = ? and status IN (%s)", m.table, strings.Join(placeholders, ","))
	result, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		return false, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	return rowsAffected > 0, nil
}

func (m *customTicketSeatModel) FindValidIdsByEventId(ctx context.Context, eventId int64) ([]int64, error) {
	// 只查 ID，减少数据传输量
	query := fmt.Sprintf("select id from %s where event_id = ? and status = 0", m.table)
	var resp []int64
	// 直接查询 ID 列，并将结果扫描到 int64 切片中
	err := m.conn.QueryRowsCtx(ctx, &resp, query, eventId)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// NewTicketSeatModel returns a usercenter for the database table.
func NewTicketSeatModel(conn sqlx.SqlConn) TicketSeatModel {
	return &customTicketSeatModel{
		defaultTicketSeatModel: newTicketSeatModel(conn),
	}
}

func (m *customTicketSeatModel) withSession(session sqlx.Session) TicketSeatModel {
	return NewTicketSeatModel(sqlx.NewSqlConnFromSession(session))
}
