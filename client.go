package mqtt

import (
	"context"
	"errors"
	"log"
	"net"
	"net/url"
	"sync"
	"time"
)

// NewClient create a new mqtt client
func NewClient(options Options) Client {
	c := &client{
		options: options,
	}
	return c
}

type client struct {
	sync.Mutex // protect conn
	conn       net.Conn
	options    Options
}

func (c *client) IsConnected() bool {
	return false
}

func (c *client) Connect(ctx context.Context) error {
	var lasterr error
	for _, s := range c.options.Servers {
		err := c.connect(ctx, s)
		if err == nil {
			return nil
		}

		log.Printf("failed to connect to %s, %s", s, err)
		lasterr = err
	}

	return lasterr
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

func (c *client) connect(ctx context.Context, url *url.URL) error {
	switch url.Scheme {
	case "tcp":
		d := net.Dialer{
			Timeout: c.options.ConnectTimeout,
		}

		conn, err := d.DialContext(ctx, "tcp", url.Host)
		if err != nil {
			return err
		}

		c.setConn(conn)
	default:
		return errors.New("unsupported protocol")
	}

	return c.cmdConnect(ctx)
}

func (c *client) setConn(conn net.Conn) {
	c.Lock()
	c.conn = conn
	c.Unlock()
}

func (c *client) cmdConnect(ctx context.Context) error {
	msg := &messageConnect{
		ClientID:  c.options.ClientID,
		Keepalive: uint16(c.options.KeepAlive / time.Second),
	}
	deadline, ok := ctx.Deadline()
	if ok {
		c.conn.SetDeadline(deadline)
	}

	_, err := c.conn.Write(msg.Encode()) // full write??
	if err != nil {
		return err
	}

	readBuf := make([]byte, 1024)
	_, err = c.conn.Read(readBuf)
	if err != nil {
		return err
	}

	log.Printf("received: %x\n", readBuf)
	return nil
}
