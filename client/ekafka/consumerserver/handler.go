package consumerserver

import (
	"context"

	"miopkg/client/ekafka"

	"github.com/segmentio/kafka-go"
)

// OnEachMessageHandler ...
type OnEachMessageHandler = func(ctx context.Context, message kafka.Message) error

// OnStartHandler ...
type OnStartHandler = func(ctx context.Context, consumer *ekafka.Consumer) error

// OnConsumerGroupStartHandler ...
type OnConsumerGroupStartHandler = func(ctx context.Context, consumerGroup *ekafka.ConsumerGroup) error
