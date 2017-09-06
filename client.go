package mqtt

import (
	"context"
	"errors"
	"log"
	"net"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"github.com/bopjiang/mqtt-client/packet"
)

type client struct {
	sync.Mutex   // TODO: protect both conn and ID ?
	conn         net.Conn
	isConnected  int64 // 0 -- disconnected, 1 -- connected
	options      Options
	nextPacketID uint16
	handler      *messageHandler

	respWaitingQueueMutex sync.Mutex
	respWaitingQueue      map[requestKey]chan interface{}

	timerResetChan chan int
	exitChan       chan struct{}
}

type requestKey struct {
	msgType byte
	id      uint16
}

// NewClient create a new mqtt client(no reconn and retry, message pending will abandoned)
func NewClient(options Options) Client {
	c := &client{
		options:          options,
		nextPacketID:     0,
		handler:          &messageHandler{},
		respWaitingQueue: make(map[requestKey]chan interface{}),
		timerResetChan:   make(chan int, 1),
	}
	return c
}

func (c *client) IsConnected() bool {
	return atomic.LoadInt64(&c.isConnected) == 1
}

func (c *client) getPacketID() uint16 {
	c.Lock()
	defer c.Unlock()
	c.nextPacketID++
	return c.nextPacketID
}

func (c *client) Connect(ctx context.Context) error {
	var lasterr error
	for _, s := range c.options.Servers {
		err := c.connect(ctx, s)
		if err == nil {
			return c.start(ctx)
		}

		log.Printf("failed to connect to %s, %s", s, err)
		lasterr = err
	}

	return lasterr
}

func (c *client) start(ctx context.Context) error {
	atomic.StoreInt64(&c.isConnected, 1)
	c.exitChan = make(chan struct{})
	go c.incomingLoop(c.conn)             // TODO: incoming return error, should notify to outgoing
	go c.outgoingLoop(c.conn, c.exitChan) // outgoing error should close the incomingLoop
	// the two loops exits, and we can start to try reconnect.
	return nil
}

func (c *client) Disconnect() error {
	msg := &packet.DisConnect{}
	c.sendPacket(msg)
	c.conn.Close()
	atomic.StoreInt64(&c.isConnected, 0)
	return nil
}

func (c *client) Publish(ctx context.Context, topic string, qos byte, retained bool, payload []byte) error {
	// retry logic?
	// reconn logic?
	return c.cmdPublish(ctx, topic, qos, false, retained, payload)
}

func (c *client) Subscribe(ctx context.Context, topic string, qos byte, callback MessageHandler) error {
	return c.cmdSubscribe(ctx, topic, qos, callback)
}

func (c *client) SubscribeMultiple(ctx context.Context, filters map[string]byte, callback MessageHandler) error {
	panic("not implemented")
}

func (c *client) Unsubscribe(ctx context.Context, topics ...string) error {
	return c.cmdUnsubscribe(ctx, topics...)
}

func (c *client) SetRoute(topic string, callback MessageHandler) {
	panic("not implemented")
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
	c.conn = conn // TODO: protection of c.conn to avoid concurrent use
	c.Unlock()
}

func (c *client) incomingLoop(conn net.Conn) error {
	var retErr error
	for {
		c.conn.SetReadDeadline(time.Now().Add(c.options.KeepAlive * 2))
		pkt, err := packet.ReadPacket(conn)
		if err != nil {
			log.Printf("failed to read packet, %s", err)
			atomic.StoreInt64(&c.isConnected, 0)
			retErr = err
			goto EXIT
		}

		switch v := pkt.(type) {
		case *packet.PubAck:
			ch, ok := c.getRequestFromQueue(packet.CtrlTypePUBACK, v.ID)
			if !ok {
				log.Printf("receive invalid puback, id=%d", v.ID)
				continue
			}
			ch <- v
		case *packet.SubAck:
			ch, ok := c.getRequestFromQueue(packet.CtrlTypeSUBACK, v.ID)
			if !ok {
				log.Printf("receive invalid suback, id=%d", v.ID)
				continue
			}
			ch <- v
		case *packet.Publish:
			if err := c.handler.Handle(&message{v.Topic, v.Payload}); err != nil {
				log.Printf("failed to process message, %s", v)
			}

			if err := c.sendPublishAck(conn, v); err != nil {
				retErr = err
				goto EXIT
			}
		case *packet.PingResp:
			// reset read timer, do nothing
		default:
			log.Printf("invalid message type, %v", v)
		}
	}

EXIT:
	conn.Close()
	return retErr
}

func (c *client) outgoingLoop(conn net.Conn, exitChan chan struct{}) {
	keepAliveTimer := time.NewTimer(c.options.KeepAlive)
	defer keepAliveTimer.Stop()
	for {
		select {
		case <-keepAliveTimer.C:
			c.sendPingReq(conn)
		case <-c.timerResetChan:
		case <-exitChan:
			goto EXIT
		}
		keepAliveTimer = time.NewTimer(c.options.KeepAlive)
	}

EXIT:
	conn.Close()
}

func (c *client) getRequestFromQueue(msgType byte, msgID uint16) (ch chan interface{}, ok bool) {
	c.respWaitingQueueMutex.Lock()
	ch, ok = c.respWaitingQueue[requestKey{msgType, msgID}]
	c.respWaitingQueueMutex.Unlock()
	return
}

func (c *client) waitPubAck(ctx context.Context, id uint16) (*packet.PubAck, error) {
	v, err := c.waitResp(ctx, packet.CtrlTypePUBACK, id)
	if err != nil {
		return nil, err
	}

	return v.(*packet.PubAck), nil
}

func (c *client) waitSubAck(ctx context.Context, id uint16) (*packet.SubAck, error) {
	v, err := c.waitResp(ctx, packet.CtrlTypeSUBACK, id)
	if err != nil {
		return nil, err
	}

	return v.(*packet.SubAck), nil
}

func (c *client) waitUnsubAck(ctx context.Context, id uint16) (*packet.UnSubAck, error) {
	v, err := c.waitResp(ctx, packet.CtrlTypeUNSUBACK, id)
	if err != nil {
		return nil, err
	}

	return v.(*packet.UnSubAck), nil
}

func (c *client) waitResp(ctx context.Context, msgType byte, id uint16) (interface{}, error) {
	respChan := make(chan interface{})
	c.respWaitingQueueMutex.Lock()
	c.respWaitingQueue[requestKey{msgType, id}] = respChan
	c.respWaitingQueueMutex.Unlock()
	select {
	case resp := <-respChan:
		return resp, nil
	case <-ctx.Done():
		log.Printf("wait resp timeout")
		return nil, ctx.Err()
	}

}

func (c *client) sendPublishAck(conn net.Conn, p *packet.Publish) error {
	return c.sendPacket(&packet.PubAck{ID: p.ID})
}

func (c *client) sendPingReq(conn net.Conn) error {
	return c.sendPacket(&packet.PingReq{})
}

func (c *client) sendPacket(p packet.PacketWriter) error {
	c.Lock()
	defer c.Unlock()
	err := p.Write(c.conn)
	if err != nil {
		atomic.StoreInt64(&c.isConnected, 0)
	}
	c.timerResetChan <- 0
	return err
}
