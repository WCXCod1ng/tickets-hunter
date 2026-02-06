package svc

import (
	"github.com/zeromicro/go-zero/zrpc"
	"tickets-hunter/app/usercenter/cmd/api/internal/config"
	"tickets-hunter/app/usercenter/cmd/rpc/usercenter"
)

type ServiceContext struct {
	Config        config.Config
	UserCenterRpc usercenter.UserCenter // 注入RPC层的RPC服务
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config: c,
		UserCenterRpc: usercenter.NewUserCenter( // 初始化RPC Client（根据Config）
			zrpc.MustNewClient(c.UserCenterRpc),
		),
	}
}
