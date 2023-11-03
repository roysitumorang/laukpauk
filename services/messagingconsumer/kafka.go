package messagingconsumer

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/IBM/sarama"
	"github.com/goccy/go-json"
	"github.com/roysitumorang/laukpauk/config"
	"github.com/roysitumorang/laukpauk/helper"
	"github.com/roysitumorang/laukpauk/router"
	"go.uber.org/zap"
)

type (
	kafkaConsumerService struct {
		consumer sarama.ConsumerGroup
	}

	client struct {
		ready   chan bool
		service *router.Service
	}
)

func NewKafkaConsumerService(brokers []string) MessagingConsumerService {
	ctxt := "MessagingConsumerKafka-NewKafkaConsumerService"
	ctx := context.Background()
	cfg := sarama.NewConfig()
	cfg.Version = sarama.V3_5_1_0
	cfg.Consumer.Return.Errors = true
	cfg.Consumer.Offsets.Initial = sarama.OffsetOldest
	groupID := "laukpauk"
	consumer, err := sarama.NewConsumerGroup(brokers, groupID, cfg)
	if err != nil {
		helper.Capture(ctx, zap.FatalLevel, fmt.Errorf("kafka: error creating consumer %s", err), ctxt, "ErrNewConsumerGroup")
	}
	service := &kafkaConsumerService{
		consumer: consumer,
	}
	return service
}

func (s *kafkaConsumerService) Consume(service *router.Service) {
	ctxt := "MessagingConsumerKafka-Consume"
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer os.Exit(0)
	topics := []string{
		config.TopicGeneral,
	}
	consumer := client{ready: make(chan bool), service: service}
	wg := &sync.WaitGroup{}
	sigterm := make(chan os.Signal, 1)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			if err := s.consumer.Consume(ctx, topics, &consumer); err != nil {
				helper.Capture(ctx, zap.FatalLevel, fmt.Errorf("kafka: error consumer %v", err), ctxt, "ErrConsume")
			}
			if ctx.Err() != nil {
				return
			}
			consumer.ready = make(chan bool)
		}
	}()
	<-consumer.ready
	helper.Log(ctx, zap.InfoLevel, "kafka: sarama consumer up and running!...", ctxt, "")
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-ctx.Done():
		helper.Log(ctx, zap.InfoLevel, "kafka: terminating via cancelled context", ctxt, "")
	case <-sigterm:
		helper.Log(ctx, zap.InfoLevel, "kafka: terminating via signal", ctxt, "")
	}
	cancel()
	wg.Wait()
	if err := s.consumer.Close(); err != nil {
		helper.Capture(ctx, zap.FatalLevel, fmt.Errorf("kafka: error closing client: %v", err), ctxt, "ErrClose")
	}
}

func (c *client) Setup(session sarama.ConsumerGroupSession) error {
	close(c.ready)
	return nil
}

func (c *client) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (c *client) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) (err error) {
	ctxt := "MessagingConsumerKafka-ConsumeClaim"
	ctx := context.Background()
	defer func() {
		if r := recover(); r != nil {
			err, ok := r.(error)
			if !ok {
				err = fmt.Errorf("%v", r)
			}
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrRecover")
		}
	}()
	for message := range claim.Messages() {
		now := time.Now()
		if json.Valid(message.Value) {
			var payload map[string]interface{}
			if err = json.Unmarshal(message.Value, &payload); err != nil {
				helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrUnmarshal")
			}
		}
		duration := time.Since(now)
		helper.Log(ctx, zap.InfoLevel, fmt.Sprintf("kafka: message on topic %s[%d]@%d: %s, consumed in %s", message.Topic, message.Partition, message.Offset, message.Value, duration.String()), ctxt, "")
		session.MarkMessage(message, "")
	}
	return nil
}
