package logic

import (
	"context"
	"database/sql"
	"tickets-hunter/app/model/user_wallet"
	"tickets-hunter/app/payment/cmd/rpc/internal/svc"
	"tickets-hunter/app/payment/cmd/rpc/payment/rpc"

	"github.com/dtm-labs/client/dtmgrpc"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type DeductLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeductLogic {
	return &DeductLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Saga 步骤1 正向操作：扣减余额
//  1. 获取 DTM 子事务屏障
//  2. 获取 go-zero 中原生的 *sql.DB 连接，用于直接使用事务
//  3. 在 DTM barrier 的事务中执行扣减余额的业务逻辑
//     注意，这里必须使用闭包注入的tx来执行数据库操作，确保所有操作都在同一个事务上下文中，
//     这样它才能和 DTM 的barrier表拦截动作同处于一个事务中，才能正确实现幂等、空回滚和悬挂处理等分布式事务特性
//  4. 处理 Barrier 的调用结果，如果是业务错误（比如余额不足），直接返回给调用方，DTM会根据这个错误类型来判断是否需要执行补偿逻辑；其他错误，记录日志并返回系统异常错误，DTM会认为这是一个系统异常，会重试这个步骤，直到成功或者达到重试上限
func (l *DeductLogic) Deduct(in *rpc.SagaPayReq) (*rpc.SagaPayResp, error) {
	l.Logger.Debugf("开始执行扣减余额逻辑，用户ID: %d, 扣减金额: %d", in.UserId, in.Amount)
	// 1. 获取 DTM 子事务屏障
	barrier, err := dtmgrpc.BarrierFromGrpc(l.ctx)
	if err != nil {
		l.Logger.Errorf("获取 DTM barrier 失败: %v", err)
		return nil, errors.WithStack(status.Error(codes.Internal, "获取 DTM barrier 失败"))
	}

	// 2. 获取 go-zero 中原生的 *sql.DB 连接，用于直接使用事务
	db, err := l.svcCtx.DB.RawDB()
	if err != nil {
		return nil, errors.WithStack(status.Error(codes.Internal, "获取数据库连接失败"))
	}

	// 3. 在 DTM barrier 的事务中执行扣减余额的业务逻辑
	err = barrier.CallWithDB(db, func(tx *sql.Tx) error {
		// 注意，这里必须使用闭包注入的tx来执行数据库操作，确保所有操作都在同一个事务上下文中
		// 这样它才能和 DTM 的barrier表拦截动作同处于一个事务中，才能正确实现幂等、空回滚和悬挂处理等分布式事务特性

		// 在事务中执行扣减余额的操作
		err = l.svcCtx.UserWalletModel.DeductWithTx(l.ctx, tx, in.UserId, in.Amount)
		if errors.Is(err, user_wallet.ErrBalanceNotEnough) {
			// 余额不足，返回特定错误，DTM会根据这个错误类型来判断是否需要执行补偿逻辑
			// 注意，这里返回的错误必须是一个 gRPC错误，并且错误码为一个特定的业务错误码（不能是 codes.Internal），比如codes.Aborted，这样 DTM 才能正确识别这是一个业务失败，需要执行补偿，而不是系统异常需要重试
			return status.Error(codes.Aborted, "余额不足，无法完成支付")
		} else if err != nil {
			// 其他错误，返回系统异常错误，DTM会认为这是一个系统异常，会重试这个步骤，直到成功或者达到重试上限
			return errors.WithStack(status.Error(codes.Internal, err.Error()))
		}

		// 扣款成功，返回 nil
		return nil
	})

	// 4. 处理 Barrier 的调用结果
	if err != nil {
		// 如果是业务错误（比如余额不足），直接返回给调用方，DTM会根据这个错误类型来判断是否需要执行补偿逻辑
		if e, ok := status.FromError(err); ok && e.Code() == codes.Aborted {
			return &rpc.SagaPayResp{
				Success: false,
				Message: "余额不足，无法完成支付",
			}, err
		}
		// 其他错误，记录日志并返回系统异常错误，DTM会认为这是一个系统异常，会重试这个步骤，直到成功或者达到重试上限
		l.Logger.Errorf("扣减余额失败: %v", err)
		return &rpc.SagaPayResp{
			Success: false,
			Message: "扣减余额失败，系统异常",
		}, errors.WithStack(status.Error(codes.Internal, err.Error()))
	}

	// 扣减成功，返回成功响应，DTM会继续执行下一个步骤
	l.Logger.Debugf("用户 %d 扣减余额 %d 成功", in.UserId, in.Amount)
	return &rpc.SagaPayResp{
		Success: true,
		Message: "扣款成功",
	}, nil
}
