package messagingproducer

import (
	"context"
	"fmt"
	"log"

	"github.com/IBM/sarama"
	"github.com/goccy/go-json"
	"github.com/roysitumorang/laukpauk/config"
	"github.com/roysitumorang/laukpauk/helper"
	"go.uber.org/zap"
)

type (
	kafkaProducerService struct {
		producer sarama.SyncProducer
	}
)

func NewKafkaProducerService(brokers []string) MessagingProducerService {
	cfg := sarama.NewConfig()
	cfg.Producer.RequiredAcks = sarama.WaitForAll
	cfg.Producer.Retry.Max = 5
	cfg.Producer.Return.Successes = true
	producer, err := sarama.NewSyncProducer(brokers, cfg)
	if err != nil {
		log.Fatalf("Error creating Kafka producer: %s", err)
	}
	service := &kafkaProducerService{
		producer: producer,
	}
	topics := []string{
		config.TopicGeneral,
	}
	for _, topic := range topics {
		_ = service.Publish(topic, map[string]interface{}{"action": "", "id": ""})
	}
	return service
}

func (s *kafkaProducerService) Publish(topic string, payloads ...map[string]interface{}) (err error) {
	ctx := context.Background()
	ctxt := "MessagingProducerKafka-Publish"
	n := len(payloads)
	if n == 0 {
		return
	}
	messages := make([]*sarama.ProducerMessage, n)
	for i, payload := range payloads {
		jsonByte, err := json.Marshal(payload)
		if err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrMarshal")
			return err
		}
		messages[i] = &sarama.ProducerMessage{
			Topic: topic,
			Value: sarama.StringEncoder(helper.ByteSlice2String(jsonByte)),
		}
	}
	if err = s.producer.SendMessages(messages); err != nil {
		helper.Capture(ctx, zap.ErrorLevel, fmt.Errorf("delivery failed: %v", err), ctxt, "ErrSendMessages")
	}
	return
}
