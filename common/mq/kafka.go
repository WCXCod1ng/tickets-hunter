package mq

import (
	"context"

	"github.com/zeromicro/go-queue/kq"
)

type GoZeroKafkaPusher struct {
	pusher *kq.Pusher
}

func NewGoZeroKafkaPusher(brokers []string, topic string) *GoZeroKafkaPusher {
	return &GoZeroKafkaPusher{
		pusher: kq.NewPusher(brokers, topic),
	}
}

func (g *GoZeroKafkaPusher) Push(ctx context.Context, data []byte) error {
	return g.pusher.Push(ctx, string(data))
}
