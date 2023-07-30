package main

import (
	"context"
	"fmt"
	"log"

	"miopkg/application"
	"miopkg/client/ekafka"
	"miopkg/client/ekafka/consumerserver"
	"miopkg/governor"
	elog "miopkg/log"

	"github.com/segmentio/kafka-go"
)

var (
	ec *ekafka.Component
)

func main() {

	eng := NewEngine()
	if err := eng.Run(); err != nil {
		elog.Panic(err.Error())
	}
}

type Engine struct {
	application.Application
}

func NewEngine() *Engine {
	eng := &Engine{}
	if err := eng.Startup(
		eng.newGovernor,
		eng.newKafkaP,
		eng.newKafkaC,
	); err != nil {
		elog.Panic("startup", elog.Any("err", err))
	}
	return eng
}

func (eng *Engine) newKafkaP() error {
	ec = ekafka.Load("kafka").Build()
	// 使用p1生产者生产消息
	produce(context.Background(), ec.Producer("p1"))
	return nil
}

func (eng *Engine) newGovernor() error {
	server := governor.StdConfig("server.governor").Build()
	return eng.Serve(server)
}

func (eng *Engine) newKafkaC() error {
	// 依赖 `ekafka` 管理 Kafka consumer
	ec = ekafka.Load("kafka").Build()
	cs := consumerserver.Load("kafkaConsumerServers.s1").Build(
		consumerserver.WithEkafka(ec),
	)

	// 用来接收、处理 `kafka-go` 和处理消息的回调产生的错误
	consumptionErrors := make(chan error)

	// 注册处理消息的回调函数
	cs.OnEachMessage(consumptionErrors, func(ctx context.Context, message kafka.Message) error {
		fmt.Printf("got a message: %s\n", string(message.Value))
		// 如果返回错误则会被转发给 `consumptionErrors`
		return nil
	})

	return eng.Serve(cs)
}

// produce 生产消息
func produce(ctx context.Context, w *ekafka.Producer) {
	// 生产3条消息
	ctx = context.WithValue(ctx, "hello", "world")
	err := w.WriteMessages(ctx,
		&ekafka.Message{Key: []byte("Key-A"), Value: []byte("Hellohahah World!22222")},
	)
	if err != nil {
		log.Fatal("failed to write messages:", err)
	}
	if err := w.Close(); err != nil {
		log.Fatal("failed to close writer:", err)
	}
}
