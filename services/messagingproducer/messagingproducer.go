package messagingproducer

import (
	"log"
	"os"
	"strings"
)

type (
	MessagingProducerService interface {
		Publish(topic string, payloads ...map[string]interface{}) (err error)
	}
)

func GetMessagingProducerService() (service MessagingProducerService) {
	switch os.Getenv("MESSAGING_SERVICE") {
	case "kafka":
		service = NewKafkaProducerService(strings.Split(os.Getenv("KAFKA_BROKERS"), ","))
	default:
		log.Fatalln("invalid messaging service provider")
	}
	return service
}
