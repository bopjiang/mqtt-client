package mqtt

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/bopjiang/mqtt-client/packet"
)

func (c *client) cmdConnect(ctx context.Context) error {
	msg := &packet.Connect{
		CleanSessionFlag: c.options.CleanSession,
		ClientID:         c.options.ClientID,
		Keepalive:        uint16(c.options.KeepAlive / time.Second),
	}
	if deadline, ok := ctx.Deadline(); ok {
		c.conn.SetDeadline(deadline)
	}

	if err := c.sendPacket(msg); err != nil {
		return err
	}

	pkt, errRead := packet.ReadPacket(c.conn)
	if errRead != nil {
		return fmt.Errorf("failed to read connack, %s", errRead)
	}

	c.conn.SetDeadline(time.Time{})
	connAck, okAck := pkt.(*packet.ConnectAck)
	if !okAck {
		return fmt.Errorf("not connack")
	}

	if connAck.ReturnCode != 0 {
		if retMsg, ok := packet.ConnackReturnCodes[connAck.ReturnCode]; ok {
			return errors.New(retMsg)
		}

		return fmt.Errorf("connack errcode=%d", connAck.ReturnCode)
	}

	log.Printf("received connack: %+v\n", connAck)
	return nil
}

// TODO: qos2 not implemented yet
func (c *client) cmdPublish(ctx context.Context, topic string,
	qos byte, dup bool, retained bool, payload []byte) error {
	msg := &packet.Publish{
		Topic:      topic,
		DupFlag:    dup,
		QosLevel:   qos,
		RetainFlag: retained,
		Payload:    payload,
		ID:         c.getPacketID(),
	}

	if err := c.sendPacket(msg); err != nil {
		return fmt.Errorf("failed to publish, %s", err)
	}

	if qos == 0 { // no need ack for QOS 0
		return nil
	}

	ack, err := c.waitPubAck(ctx, msg.ID)
	if err != nil {
		return err
	}

	// It MUST send PUBACK packets in the order in which the corresponding PUBLISH packets were received (QoS 1 messages) [MQTT-4.6.0-2]
	log.Printf("received puback: %+v, publish id=%d\n", ack, msg.ID)

	if ack.ID != msg.ID {
		return fmt.Errorf("puback id error")
	}

	return nil
}

func (c *client) cmdSubscribe(ctx context.Context, topic string, qos byte, callback MessageHandler) error {
	msg := &packet.Subscribe{
		ID:          c.getPacketID(),
		TopicFilter: []string{topic},
		QosLevel:    []byte{qos},
	}

	if err := c.sendPacket(msg); err != nil {
		return err
	}

	ack, err := c.waitSubAck(ctx, msg.ID)
	if err != nil {
		return err
	}

	if len(ack.RetCode) != len(msg.QosLevel) {
		errors.New("return code number does not match")
	}

	if ack.RetCode[0] != 0 { // TODO: qos1 returncode=1
		errors.New("sub error")
	}

	c.handler.Register(topic, qos, callback)
	return nil
}

func (c *client) cmdUnsubscribe(ctx context.Context, topics ...string) error {
	msg := &packet.UnSubscribe{
		ID: c.getPacketID(),
	}

	if err := c.sendPacket(msg); err != nil {
		return err
	}

	_, err := c.waitUnsubAck(ctx, msg.ID)
	if err != nil {
		return err
	}

	return nil
}
