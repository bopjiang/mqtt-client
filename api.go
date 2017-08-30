package mqtt

import "context"

// Client defines the interface of this library
type Client interface {
	// IsConnected returns the status of the client
	IsConnected() bool

	// all the function below could block, use Context to cancel or timetout.
	Connect(ctx context.Context) error

	// Disconnect close client connection with a waiting time
	Disconnect(ctx context.Context) error

	// Pushlish push message to topic
	Publish(ctx context.Context, topic string, qos byte, retained bool, payload []byte) error

	// Subscribe subscribes a single topic. Callback could be nil
	Subscribe(ctx context.Context, topic string, qos byte, callback MessageHandler) error

	// SubscribeMultiple subscribes mutiple topics. Callback could be nil
	SubscribeMultiple(ctx context.Context, filters map[string]byte, callback MessageHandler) error

	// Unsubscribe unsubscribes mutiple topics
	Unsubscribe(ctx context.Context, topics ...string) error

	// SetRoute set the callback of topic, and overide the callback setting in Subscribe or SubscribeMultiple
	SetRoute(topic string, callback MessageHandler)
}

// MessageHandler is a callback type which can be set to be
// executed upon the arrival of messages published to topics
// to which the client is subscribed.
type MessageHandler func(Client, Message)

// Message define the interface of mqtt message(no QoS, retained info here)
type Message interface {
	Topic() string
	Payload() []byte
}
