package mq

import (
	"context"
	"sync"
)

type Handler func(context.Context, []byte) error

type Dispatcher struct {
	handlers map[string]Handler
	mu       sync.RWMutex
}

func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		handlers: make(map[string]Handler),
	}
}

// 注册 handler（业务调用）
func (d *Dispatcher) Register(eventType string, handler Handler) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.handlers[eventType] = handler
}

// 分发消息（内部使用）
func (d *Dispatcher) Dispatch(ctx context.Context, eventType string, data []byte) error {
	d.mu.RLock()
	handler, ok := d.handlers[eventType]
	d.mu.RUnlock()

	if !ok {
		return nil // 或记录日志
	}

	return handler(ctx, data)
}
