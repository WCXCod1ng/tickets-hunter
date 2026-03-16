package logic

import (
	"context"
	"database/sql"
	"tickets-hunter/app/payment/cmd/rpc/internal/svc"
	"tickets-hunter/app/payment/cmd/rpc/payment/rpc"

	"github.com/dtm-labs/client/dtmgrpc"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type RefundLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRefundLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RefundLogic {
	return &RefundLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Saga 步骤2 补偿操作：退还余额
func (l *RefundLogic) Refund(in *rpc.SagaPayReq) (*rpc.SagaPayResp, error) {
	l.Logger.Debugf("开始执行退款补偿逻辑，userId: %d, amount: %d", in.UserId, in.Amount)
	barrier, err := dtmgrpc.BarrierFromGrpc(l.ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	db, err := l.svcCtx.DB.RawDB()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	err = barrier.CallWithDB(db, func(tx *sql.Tx) error {
		// 退款补偿：直接把钱加回去
		err := l.svcCtx.UserWalletModel.RefundWithTx(l.ctx, tx, in.UserId, in.Amount)

		// 哪怕退款失败，只要抛出 err，DTM 的定时任务就会一直疯狂重试这个补偿动作，直到成功为止，彻底保证最终一致性（资损为0）！
		return err
	})

	if err != nil {
		l.Logger.Errorf("退款补偿失败，DTM将稍后重试。userId: %d, amount: %d, error: %v", in.UserId, in.Amount, err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &rpc.SagaPayResp{Success: true, Message: "退款补偿成功"}, nil
}
