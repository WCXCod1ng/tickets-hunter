package user_wallet

import (
	"context"
	"database/sql"
	"fmt"

	errors2 "github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ UserWalletModel = (*customUserWalletModel)(nil)

var ErrBalanceNotEnough = errors2.New("余额不足")

type (
	// UserWalletModel is an interface to be customized, add more methods here,
	// and implement the added methods in customUserWalletModel.
	UserWalletModel interface {
		userWalletModel
		withSession(session sqlx.Session) UserWalletModel

		// 扣款
		Deduct(ctx context.Context, userId int64, amount int64) error

		// 带事务的扣款
		DeductWithTx(ctx context.Context, tx *sql.Tx, userId int64, amount int64) error

		// 退款
		Refund(ctx context.Context, userId int64, amount int64) error

		// 带事务的退款
		RefundWithTx(ctx context.Context, tx *sql.Tx, userId int64, amount int64) error
	}

	customUserWalletModel struct {
		*defaultUserWalletModel
	}
)

func (m *customUserWalletModel) Deduct(ctx context.Context, userId int64, amount int64) error {
	query := "UPDATE user_wallet SET balance = balance - ? WHERE user_id = ? AND balance >= ?"
	result, err := m.conn.ExecCtx(ctx, query, amount, userId, amount)
	if err != nil {
		return errors2.WithStack(err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors2.WithStack(err)
	}
	if rowsAffected == 0 {
		return ErrBalanceNotEnough
	}
	return nil
}

func (m *customUserWalletModel) DeductWithTx(ctx context.Context, tx *sql.Tx, userId int64, amount int64) error {
	query := "UPDATE user_wallet SET balance = balance - ? WHERE user_id = ? AND balance >= ?"
	res, err := tx.ExecContext(ctx, query, amount, userId, amount)
	if err != nil {
		return errors2.WithStack(err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return errors2.WithStack(err)
	}

	if rowsAffected == 0 {
		return ErrBalanceNotEnough
	}

	return nil
}

func (m *customUserWalletModel) Refund(ctx context.Context, userId int64, amount int64) error {
	query := "UPDATE user_wallet SET balance = balance + ? WHERE user_id = ?"
	result, err := m.conn.ExecCtx(ctx, query, amount, userId)
	if err != nil {
		return errors2.WithStack(err)
	}
	if rowsAffected, err := result.RowsAffected(); err != nil || rowsAffected == 0 {
		return errors2.WithStack(fmt.Errorf("退款失败，userId: %d, amount: %d, err: %v, rowsAffected: %d", userId, amount, err, rowsAffected))
	}
	return nil
}

func (m *customUserWalletModel) RefundWithTx(ctx context.Context, tx *sql.Tx, userId int64, amount int64) error {
	query := "UPDATE user_wallet SET balance = balance + ? WHERE user_id = ?"
	result, err := tx.ExecContext(ctx, query, amount, userId)
	if err != nil {
		return errors2.WithStack(err)
	}
	if rowsAffected, err := result.RowsAffected(); err != nil || rowsAffected == 0 {
		return errors2.WithStack(fmt.Errorf("退款失败，userId: %d, amount: %d, err: %v, rowsAffected: %d", userId, amount, err, rowsAffected))
	}
	return nil
}

// NewUserWalletModel returns a model for the database table.
func NewUserWalletModel(conn sqlx.SqlConn) UserWalletModel {
	return &customUserWalletModel{
		defaultUserWalletModel: newUserWalletModel(conn),
	}
}

func (m *customUserWalletModel) withSession(session sqlx.Session) UserWalletModel {
	return NewUserWalletModel(sqlx.NewSqlConnFromSession(session))
}
