package main

import (
	"context"
	"fmt"
	"log"
	"miopkg/application"
	"miopkg/client/ekafka"
	ilog "miopkg/log"
	etrace "miopkg/trace"
)

func main() {
	eng := NewEngine()
	if err := eng.Run(); err != nil {
		log.Panic(err.Error())
	}
}

type Engine struct {
	application.Application
}

func NewEngine() *Engine {
	eng := &Engine{}
	if err := eng.Startup(
		eng.newkafka,
	); err != nil {
		ilog.Panic("startup", ilog.Any("err", err))
	}
	return eng
}

func (eng *Engine) newkafka() error {
	ctx := context.Background()
	// 初始化ekafka组件
	cmp := ekafka.Load("kafka").Build()
	// 使用p1生产者生产消息
	produce(ctx, cmp.Producer("p1"))

	//md.ForeachKey(func(key, val string) error {
	//	fmt.Println(key)
	//	fmt.Println(val)
	//	return nil
	//})

	// 使用c1消费者消费消息
	consume(cmp.Consumer("c1"))
	return nil
}

// produce 生产消息
func produce(ctx context.Context, w *ekafka.Producer) {
	// 生产3条消息
	ctx = context.WithValue(ctx, "hello", "world")
	err := w.WriteMessages(ctx,
		&ekafka.Message{Key: []byte("Key-A"), Value: []byte("Hello traceinterceptor")},
	)
	if err != nil {
		log.Fatal("failed to write messages:", err)
	}
	if err := w.Close(); err != nil {
		log.Fatal("failed to close writer:", err)
	}
}

// consume 使用consumer/consumerGroup消费消息
func consume(r *ekafka.Consumer) {
	ctx := context.Background()
	for {
		// ReadMessage 再收到下一个Message时，会阻塞
		msg, ctxOutput, err := r.ReadMessage(ctx)
		if err != nil {
			panic("could not read message " + err.Error())
		}
		fmt.Printf("etrace.ExtractTraceID(ctxOutput)--------------->"+"%+v\n", etrace.ExtractTraceID(ctxOutput))
		// 打印消息
		fmt.Println("received headers: ", msg.Headers)
		fmt.Println("received: ", string(msg.Value))
		err = r.CommitMessages(ctxOutput, &msg)
		if err != nil {
			log.Printf("fail to commit msg:%v", err)
		}
	}
}