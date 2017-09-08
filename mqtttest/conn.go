package mqtttest

import (
	"net"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/bopjiang/mqtt-client/packet"
)

type protocol interface {
	Serve()
}

type mqttConn struct {
	net.Conn

	t            *testing.T
	disconnected int64

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

func (c *mqttConn) Close() {
	if atomic.LoadInt64(&c.disconnected) == 1 {
		c.wg.Wait()
		return
	}

	close(c.connExitCh)
	c.Close()
	c.wg.Wait()
}

func (c *mqttConn) incomingLoop() (err error) {
	defer c.wg.Done()
	for {
		pkt, readErr := packet.ReadPacket(c)
		if readErr != nil {
			atomic.StoreInt64(&c.disconnected, 1)
			err = readErr
			goto EXIT
		}

		_ = pkt
	}

EXIT:
	close(c.connExitCh)
	return
}

func (c *mqttConn) outgoingLoop() error {
	return nil
}
