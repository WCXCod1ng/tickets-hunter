package main

import (
	"context"
	"flag"
	"tickets-hunter/app/order/cmd/orderJob/internal/config"
	"tickets-hunter/app/order/cmd/orderJob/internal/logic"
	"tickets-hunter/app/order/cmd/orderJob/internal/svc"
	"tickets-hunter/common/job"

	"github.com/zeromicro/go-zero/core/conf"
)

var configFile = flag.String("f", "etc/orderjob.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	svcCtx := svc.NewServiceContext(c)

	// 注册任务
	registry := job.NewRegistry()

	registry.Register(logic.NewProcessDelayTaskJob(svcCtx))

	ctx := context.Background()
	// 启动任务调度器
	registry.BlockRunAll(ctx)
}
