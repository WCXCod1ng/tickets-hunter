package main

import (
	"flag"
	"fmt"
	"tickets-hunter/common/interceptor"

	"tickets-hunter/app/usercenter/cmd/rpc/internal/config"
	"tickets-hunter/app/usercenter/cmd/rpc/internal/server"
	"tickets-hunter/app/usercenter/cmd/rpc/internal/svc"
	"tickets-hunter/app/usercenter/cmd/rpc/usercenter/rpc"

	grpc_validator "github.com/grpc-ecosystem/go-grpc-middleware/validator"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	_ "github.com/go-sql-driver/mysql"
)

var configFile = flag.String("f", "etc/usercenter.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	ctx := svc.NewServiceContext(c)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		rpc.RegisterUserCenterServer(grpcServer, server.NewUserCenterServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	// 添加中间件
	s.AddUnaryInterceptors(
		// 添加全局错误处理中间件
		interceptor.ServerErrorInterceptor,
		// 添加全局参数校验中间件。注意，必须在全局错误处理中间件之后添加，否则参数校验失败时错误信息将无法正确传递到错误处理中间件中。
		grpc_validator.UnaryServerInterceptor(),
	)

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
