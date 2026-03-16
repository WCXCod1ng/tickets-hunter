package utils

import (
	"errors"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/zrpc"
)

// 构建 RPC 连接目标字符串，优先使用 conf.Target，如果没有则根据 etcd 配置构建
func BuildTarget(conf zrpc.RpcClientConf) (string, error) {
	if conf.Target != "" {
		return conf.Target, nil
	}

	if len(conf.Etcd.Hosts) > 0 && conf.Etcd.Key != "" {
		return fmt.Sprintf(
			"etcd://%s/%s",
			strings.Join(conf.Etcd.Hosts, ","),
			conf.Etcd.Key,
		), nil
	}

	return "", errors.New("rpc client conf has no target or etcd config")
}
