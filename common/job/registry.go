package job

import (
	"context"
	"sync"

	"github.com/zeromicro/go-zero/core/logx"
)

type Registry struct {
	tasks []Job
	mu    sync.Mutex
}

func NewRegistry() *Registry {
	return &Registry{
		tasks: make([]Job, 0),
	}
}

// Register 注册一个任务
func (r *Registry) Register(t Job) {
	if t == nil {
		panic("task is nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.tasks = append(r.tasks, t)
}

// RunAll 启动所有任务（每个任务一个 goroutine）
func (r *Registry) RunAll(ctx context.Context) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, t := range r.tasks {
		task := t // 防止闭包问题

		logx.Infof("start task: %s", task.Name())

		go func() {
			defer func() {
				if err := recover(); err != nil {
					logx.Errorf("task panic [%s]: %v", task.Name(), err)
				}
			}()

			task.Run(ctx)
		}()
	}
}

// 启动所有任务并阻塞当前 goroutine，适用于 main 函数中调用
func (r *Registry) BlockRunAll(ctx context.Context) {
	r.RunAll(ctx)
	select {}
}
