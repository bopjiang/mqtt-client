package mqtt

import "log"

type messageHandlerInterface interface {
	Register(topicFilters []string, qos []byte, callback MessageHandler)
	Handle(client Client, message Message)
}

type filter struct {
	topicFilter string
	qos         byte
}

type messageHandler struct {
	handlers map[filter]MessageHandler
}

func (h *messageHandler) Register(topicFilters string, qos byte, callback MessageHandler) {
	log.Printf("register topicfilters=%s\n", topicFilters)
}

func (h *messageHandler) Handle(client Client, message Message) {
	log.Printf("message received [%s]:  %s\n", message.Topic(), message.Payload())
}
