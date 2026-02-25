// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package main

import (
	"flag"
	"fmt"
	"tickets-hunter/common/result"
	"tickets-hunter/common/validator"

	"tickets-hunter/app/ticket/cmd/api/internal/config"
	"tickets-hunter/app/ticket/cmd/api/internal/handler"
	"tickets-hunter/app/ticket/cmd/api/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
)

var configFile = flag.String("f", "etc/ticket-api.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	// 注册自定义参数校验器，必须在创建服务器之前调用
	httpx.SetValidator(validator.New())

	// 注册全局错误处理器
	result.RegisterErrorHandler()

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
