package idgen

import (
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/zeromicro/go-zero/core/logx"
)

// 创建一个 Snowflake 节点
func CreateSnowFlakeNode(startTime string, workerId int64) *snowflake.Node {
	// 解析配置文件中的时间字符串
	t, err := time.Parse("2006-01-02", startTime)
	if err != nil {
		// 如果格式不对，建议直接 Panic，强制修正配置
		logx.Errorf("Snowflake StartTime parse error: %v, use default", err)
		panic("Invalid Snowflake StartTime format, use YYYY-MM-DD")
	}

	// 设置全局 Epoch (关键步骤)
	// 注意：snowflake.Epoch 是一个全局变量，必须在 NewNode 之前设置
	snowflake.Epoch = t.UnixMilli()

	// 初始化 snowflake 节点
	node, err := snowflake.NewNode(workerId)
	if err != nil {
		// 如果初始化失败（例如 WorkerId 超出范围），记录日志并 panic
		logx.Errorf("init snowflake node failed: %v", err)
		panic(err)
	}

	return node
}
