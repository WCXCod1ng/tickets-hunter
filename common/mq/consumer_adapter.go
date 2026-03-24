package mq

import "context"

type MQConsumerAdapter struct {
	consumer *Consumer
}

func NewMQConsumerAdapter(c *Consumer) *MQConsumerAdapter {
	return &MQConsumerAdapter{
		consumer: c,
	}
}

// 实现 go-zero 的接口
func (a *MQConsumerAdapter) Consume(ctx context.Context, key, value string) error {
	a.consumer.Consume(ctx, []byte(value))
	return nil
}
