package job

import "context"

// 所有任务必须实现的接口
type Job interface {
	Name() string // 任务名，用于日志
	Run(ctx context.Context)
}
