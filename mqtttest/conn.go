package mqtttest

import (
	"io"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/openim/mqtt-client/packet"
)

type protocol interface {
	Serve()
	SetTimeout(time.Duration)
}

type mqttConn struct {
	net.Conn
	connLock sync.Mutex

	t            *testing.T
	disconnected int64
	timeout      time.Duration // read timeout

	serverExitCh chan struct{}
	connExitCh   chan struct{}
	wg           sync.WaitGroup
}

func newMQTTConn(t *testing.T, conn net.Conn, exitCh chan struct{}) protocol {
	return &mqttConn{
		Conn:         conn,
		t:            t,
		serverExitCh: exitCh,
		connExitCh:   make(chan struct{}),
	}
}

func (c *mqttConn) Serve() {
	c.wg.Add(2)
	go c.outgoingLoop()
	go c.incomingLoop()
}

func (c *mqttConn) SetTimeout(t time.Duration) {
	c.timeout = t
}

func (c *mqttConn) Close() {
	if atomic.LoadInt64(&c.disconnected) == 1 {
		c.wg.Wait()
		return
	}

	c.Close()
	c.wg.Wait()
}

func (c *mqttConn) incomingLoop() (err error) {
	defer c.wg.Done()
	for {
		c.SetReadDeadline(time.Now().Add(c.timeout))
		pkt, readErr := packet.ReadPacket(c)
		if readErr != nil {
			atomic.StoreInt64(&c.disconnected, 1)
			err = readErr
			goto EXIT
		}

		switch v := pkt.(type) {
		case *packet.PingReq:
			log.Printf("received ping req")
			ack := &packet.PingResp{}
			if sendErr := c.Send(ack); sendErr != nil {
				err = sendErr
				goto EXIT
			}
		case *packet.DisConnect:
			log.Printf("received disconnected")
			goto EXIT
		case *packet.Subscribe:
			log.Printf("received subscribe")
			ack := &packet.SubAck{
				ID:      v.ID,
				RetCode: make([]byte, len(v.TopicFilter)), // TODO
			}

			for i, _ := range v.TopicFilter {
				ack.RetCode[i] = 0
			}

			if sendErr := c.Send(ack); sendErr != nil {
				err = sendErr
				goto EXIT
			}

		case *packet.Publish:
			ack := &packet.PubAck{
				ID: v.ID,
			}

			if sendErr := c.Send(ack); sendErr != nil {
				err = sendErr
				goto EXIT
			}

		default:
			c.Errorf("message not processed, %v", v)
		}
	}

EXIT:
	close(c.connExitCh)
	return
}

func (c *mqttConn) outgoingLoop() error {
	return nil
}

type writepacket interface {
	Write(io.Writer) error
}

func (c *mqttConn) Send(p writepacket) error {
	c.connLock.Lock()
	defer c.connLock.Unlock()
	return p.Write(c)
}

func (c *mqttConn) Errorf(format string, args ...interface{}) {
	c.t.Errorf(format, args...)
}
