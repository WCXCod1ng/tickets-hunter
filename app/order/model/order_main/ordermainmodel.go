package order_main

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

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
	}

	customOrderMainModel struct {
		*defaultOrderMainModel
	}

	orderMainDB struct {
		//Id         int64        `db:"id"`
		OrderSn string `db:"order_sn"` // 订单流水号(雪花算法生成)
		//UserId     int64        `db:"user_id"`  // 用户ID
		EventId int64 `db:"event_id"` // 场次ID
		SeatId  int64 `db:"seat_id"`  // 座位ID
		Amount  int64 `db:"amount"`   // 订单实付金额，以分为单位
		Status  int64 `db:"status"`   // 状态: 10待支付, 20已支付(待出票), 30已出票(已完成)，40超时关闭，50已退款
		//PayTime    sql.NullTime `db:"pay_time"` // 实际支付时间
		ExpireTime sql.NullTime `db:"expire_time"`
		CreateTime sql.NullTime `db:"create_time"`
		//UpdateTime sql.NullTime `db:"update_time"`
	}
)

func (m *customOrderMainModel) FindByOrderSnAndUserId(ctx context.Context, orderSn string, userId int64) (*OrderMain, error) {
	columnsStr := strings.Join([]string{"`order_sn`", "`event_id`", "`seat_id`", "`amount`", "`status`", "`expire_time`", "`create_time`"}, ",")
	query := fmt.Sprintf("select %s from %s where order_sn = ? and user_id = ?", columnsStr, m.table)
	var dbRow orderMainDB
	err := m.conn.QueryRowCtx(ctx, &dbRow, query, orderSn, userId)
	switch err {
	case nil:
		orderMain := &OrderMain{
			//Id:         dbRow.Id,
			OrderSn: dbRow.OrderSn,
			//UserId:     dbRow.UserId,
			EventId:    dbRow.EventId,
			SeatId:     dbRow.SeatId,
			Amount:     dbRow.Amount,
			Status:     dbRow.Status,
			ExpireTime: dbRow.ExpireTime.Time,
			//PayTime:    dbRow.PayTime,
			CreateTime: dbRow.CreateTime.Time,
			UpdateTime: dbRow.CreateTime.Time,
		}
		return orderMain, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

// NewOrderMainModel returns a model for the database table.
func NewOrderMainModel(conn sqlx.SqlConn) OrderMainModel {
	return &customOrderMainModel{
		defaultOrderMainModel: newOrderMainModel(conn),
	}
}

func (m *customOrderMainModel) withSession(session sqlx.Session) OrderMainModel {
	return NewOrderMainModel(sqlx.NewSqlConnFromSession(session))
}
