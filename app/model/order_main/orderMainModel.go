package order_main

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ OrderMainModel = (*customOrderMainModel)(nil)

type (
	// OrderMainModel is an interface to be customized, add more methods here,
	// and implement the added methods in customOrderMainModel.
	OrderMainModel interface {
		orderMainModel
		withSession(session sqlx.Session) OrderMainModel

		// 根据订单号和用户id查询订单信息
		FindByOrderSnAndUserId(ctx context.Context, orderSn string, userId int64) (*OrderMain, error)

		// 根据订单号和旧状态更新订单状态，主要用于订单超时未支付的场景
		UpdateStatusByOrderSnAndStatus(ctx context.Context, orderSn string, oldStatus int64, newStatus int64) (bool, error)

		// 带事务的根据订单号和旧状态更新订单状态，主要用于订单支付成功后更新订单状态的场景
		UpdateStatusByOrderSnAndStatusWithTx(ctx context.Context, tx *sql.Tx, orderSn string, oldStatus int64, newStatus int64) (bool, error)

		UpdateStatusByOrderSnAndNotStatusWithTx(ctx context.Context, tx *sql.Tx, orderSn string, notStatus int64, newStatus int64) (bool, error)

		FindByStatusAndExpireTimeLessThan(ctx context.Context, orderStatus int64, expireTime time.Time) ([]*OrderMain, error)
	}

	customOrderMainModel struct {
		*defaultOrderMainModel
	}
)

func (m *customOrderMainModel) FindByOrderSnAndUserId(ctx context.Context, orderSn string, userId int64) (*OrderMain, error) {
	columnsStr := strings.Join([]string{"`order_sn`", "`event_id`", "`seat_id`", "`amount`", "`status`", "`expire_time`", "`create_time`"}, ",")
	query := fmt.Sprintf("select %s from %s where order_sn = ? and user_id = ?", columnsStr, m.table)
	var dbRow OrderMain
	err := m.conn.QueryRowPartialCtx(ctx, &dbRow, query, orderSn, userId)
	switch err {
	case nil:
		return &dbRow, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customOrderMainModel) UpdateStatusByOrderSnAndStatus(ctx context.Context, orderSn string, oldStatus int64, newStatus int64) (bool, error) {
	query := fmt.Sprintf("update %s set `status` = ? where order_sn = ? and `status` = ?", m.table)
	result, err := m.conn.ExecCtx(ctx, query, newStatus, orderSn, oldStatus)
	if err != nil {
		return false, err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return rowsAffected > 0, nil
}

func (m *customOrderMainModel) UpdateStatusByOrderSnAndStatusWithTx(ctx context.Context, tx *sql.Tx, orderSn string, oldStatus int64, newStatus int64) (bool, error) {
	query := fmt.Sprintf("update %s set `status` = ? where order_sn = ? and `status` = ?", m.table)
	result, err := tx.ExecContext(ctx, query, newStatus, orderSn, oldStatus)
	if err != nil {
		return false, err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return rowsAffected > 0, nil
}

func (m *customOrderMainModel) UpdateStatusByOrderSnAndNotStatusWithTx(ctx context.Context, tx *sql.Tx, orderSn string, notStatus int64, newStatus int64) (bool, error) {
	query := fmt.Sprintf("update %s set `status` = ? where order_sn = ? and `status` != ?", m.table)
	result, err := tx.ExecContext(ctx, query, newStatus, orderSn, notStatus)
	if err != nil {
		return false, err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return rowsAffected > 0, nil
}

func (m *customOrderMainModel) FindByStatusAndExpireTimeLessThan(ctx context.Context, orderStatus int64, now time.Time) ([]*OrderMain, error) {
	query := fmt.Sprintf("select `order_sn`, `seat_id`, `seat_index`, `section`, `event_id` from %s where `status` = ? and `expire_time` <= DATE_SUB(?, INTERVAL 5 MINUTE) limit 500", m.table)
	var res []*OrderMain
	err := m.conn.QueryRowsPartialCtx(ctx, &res, query, orderStatus, now)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// NewOrderMainModel returns a usercenter for the database table.
func NewOrderMainModel(conn sqlx.SqlConn) OrderMainModel {
	return &customOrderMainModel{
		defaultOrderMainModel: newOrderMainModel(conn),
	}
}

func (m *customOrderMainModel) withSession(session sqlx.Session) OrderMainModel {
	return NewOrderMainModel(sqlx.NewSqlConnFromSession(session))
}
