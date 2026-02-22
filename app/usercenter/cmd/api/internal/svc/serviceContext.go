package svc

import (
	"tickets-hunter/app/usercenter/cmd/api/internal/config"
	"tickets-hunter/app/usercenter/cmd/rpc/usercenter"
	"tickets-hunter/common/serialize"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config        config.Config
	UserCenterRpc usercenter.UserCenter // 注入RPC层的RPC服务
	Redis         *redis.Redis
	Serializer    serialize.Serializer
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config: c,
		UserCenterRpc: usercenter.NewUserCenter( // 初始化RPC Client（根据Config）
			zrpc.MustNewClient(c.UserCenterRpc),
		),
		Redis:      redis.MustNewRedis(c.Redis),
		Serializer: serialize.JSONSerializer{}, // 初始化序列化器，这里使用JSON作为示例（方便调试），实际应当使用更高效的序列化器，如Protobuf或者MsgPack
	}
}
