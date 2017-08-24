package mqtt

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/bopjiang/mqtt-client/packet"
)

func (c *client) cmdConnect(ctx context.Context) error {
	msg := &packet.MessageConnect{
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

	pkt, errRead := packet.ReadPacket(c.conn)
	if errRead != nil {
		return fmt.Errorf("failed to read connack, %s", errRead)
	}

	connAck, okAck := pkt.(packet.MessageConnectAck)
	if okAck {
		return fmt.Errorf("not connack")
	}

	log.Printf("received: %+v\n", connAck)
	return nil
}
