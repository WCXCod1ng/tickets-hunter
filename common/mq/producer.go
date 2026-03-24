package mq

import (
	"context"
	"tickets-hunter/common/msg/mq_msg"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/proto"
)

type Producer struct {
	pusher KafkaPusher // 抽象接口（方便替换 Kafka）
}

// 抽象 Kafka（方便测试/扩展）
type KafkaPusher interface {
	Push(ctx context.Context, data []byte) error
}

func NewProducer(pusher KafkaPusher) *Producer {
	return &Producer{
		pusher: pusher,
	}
}

// 发送业务消息（统一入口）
func (p *Producer) Send(ctx context.Context, eventType string, msg proto.Message) error {
	bizData, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	wrapper := &mq_msg.MQMessage{
		Type:    eventType,
		Version: 1,
		Data:    bizData,
	}

	data, err := proto.Marshal(wrapper)
	if err != nil {
		return err
	}

	if err := p.pusher.Push(ctx, data); err != nil {
		logx.Error("kafka push failed:", err)
		return err
	}

	return nil
}
