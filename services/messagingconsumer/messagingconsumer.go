package messagingconsumer

import (
	"log"
	"os"
	"strings"

	"github.com/roysitumorang/laukpauk/router"
)

type (
	MessagingConsumerService interface {
		Consume(service *router.Service)
	}
)

func GetMessagingConsumerService() (service MessagingConsumerService) {
	switch os.Getenv("MESSAGING_SERVICE") {
	case "kafka":
		service = NewKafkaConsumerService(strings.Split(os.Getenv("KAFKA_BROKERS"), ","))
	default:
		log.Fatalln("invalid messaging service provider")
	}
	return service
}
