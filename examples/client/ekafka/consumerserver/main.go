package main

import (
	"context"
	"fmt"
	"strings"

	"miopkg/application"
	"miopkg/client/ekafka"
	"miopkg/client/ekafka/consumerserver"
	econf "miopkg/conf"
	"miopkg/governor"
	elog "miopkg/log"

	"github.com/BurntSushi/toml"
	"github.com/segmentio/kafka-go"
)

func main() {

	conf := `
	[kafka]
	debug=true
	brokers=["localhost:9092"]
	[kafka.client]
        timeout="3s"
	[kafka.producers.p1]        # 定义了名字为p1的producer
		topic="comment-events"  # 指定生产消息的topic

	[kafka.consumers.c1]        # 定义了名字为c1的consumer
		topic="comment-events"  # 指定消费的topic
		groupID="group-1"       # 如果配置了groupID,将初始化为consumerGroup
	[kafkaConsumerServers.s1]
	debug=true
	consumerName="c1"
`

	// 加载配置文件
	err := econf.LoadFromReader(strings.NewReader(conf), toml.Unmarshal)
	if err != nil {
		panic("LoadFromReader fail," + err.Error())
	}

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
		eng.newkafkaCS,
	); err != nil {
		elog.Panic("startup", elog.Any("err", err))
	}
	return eng
}

func (eng *Engine) newGovernor() error {
	server := governor.StdConfig("server.governor").Build()
	return eng.Serve(server)
}

func (eng *Engine) newkafkaCS() error {
	// 初始化 Consumer Server
	// 依赖 `ekafka` 管理 Kafka consumer
	fmt.Printf("newkafkaCS")
	ec := ekafka.Load("kafka").Build()
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
