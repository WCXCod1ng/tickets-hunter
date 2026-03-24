package mq

import (
	"context"
	"tickets-hunter/common/msg/mq_msg"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/proto"
)

type Consumer struct {
	Dispatcher *Dispatcher
}

func NewConsumer(dispatcher *Dispatcher) *Consumer {
	return &Consumer{
		Dispatcher: dispatcher,
	}
}

// 供 Kafka 调用（统一入口）
func (c *Consumer) Consume(ctx context.Context, data []byte) error {
	var msg mq_msg.MQMessage
	if err := proto.Unmarshal(data, &msg); err != nil {
		logx.Error("decode MQMessage failed:", err)
		return err
	}

	if err := c.Dispatcher.Dispatch(ctx, msg.Type, msg.Data); err != nil {
		logx.Error("dispatch failed:", err)
		return err
	}

	return nil
}
