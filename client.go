package mqtt

import (
	"context"
	"net"
)

// NewClient create a new mqtt client
func NewClient(options Options) (Client, error) {
	c := &client{
		options: options,
	}
	return c, nil
}

type client struct {
	conn    net.Conn
	options Options
}

func (c *client) IsConnected() bool {
	return false
}

func (c *client) Connect(ctx context.Context) error {
	return nil
}

func (c *client) Disconnect(ctx context.Context) error {
	return nil
}

func (c *client) Publish(ctx context.Context, topic string, qos byte, retained bool, payload interface{}) error {
	return nil
}

func (c *client) Subscribe(ctx context.Context, topic string, qos byte, callback MessageHandler) error {
	return nil
}

func (c *client) SubscribeMultiple(ctx context.Context, filters map[string]byte, callback MessageHandler) error {
	return nil
}

func (c *client) Unsubscribe(ctx context.Context, topics ...string) error {
	return nil
}

func (c *client) SetRoute(topic string, callback MessageHandler) {
	return
}
