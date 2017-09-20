package packet

import (
	"bytes"
	"encoding/binary"
	"io"
)

type Publish struct {
	FixedHeader
	Topic      string
	DupFlag    bool
	QosLevel   byte
	RetainFlag bool
	Payload    []byte
	ID         uint16 // Packet Identifier
}

const (
	Qos0 = byte(0)
	Qos1 = byte(1)
	Qos2 = byte(2)
)

const (
	publishOffsetRetain = 0
	publishOffsetQos    = 1
	publishOffsetDup    = 3
)

func (msg *Publish) Read(r io.Reader) error {
	msg.RetainFlag = (msg.Flag | 1<<publishOffsetRetain) == 1
	msg.QosLevel = (msg.Flag >> publishOffsetQos) & 0x03
	msg.DupFlag = (msg.Flag | 1<<publishOffsetDup) == 1

	buf := make([]byte, msg.RemainingLen)
	if _, err := io.ReadFull(r, buf); err != nil {
		return err
	}

	topicLen := binary.BigEndian.Uint16(buf[:2])
	msg.Topic = string(buf[2 : 2+topicLen])
	buf = buf[2+topicLen:]
	if msg.QosLevel != Qos0 {
		msg.ID = binary.BigEndian.Uint16(buf[:2])
	}

	msg.Payload = buf[2:]
	//log.Printf("received in pub %+v\n", msg)
	return nil
}

func (msg *Publish) Write(w io.Writer) error {
	var (
		remainingLength int
		fixHeaderflag   byte
	)

	remainingLength += (2 + len(msg.Topic)) // Topic Name
	if msg.QosLevel != Qos0 {
		remainingLength += 2 // Packet Identifier
	}
	remainingLength += len(msg.Payload)

	if msg.RetainFlag {
		fixHeaderflag |= 1 << publishOffsetRetain
	}
	if msg.QosLevel > 0 {
		fixHeaderflag |= msg.QosLevel << publishOffsetQos
	}
	if msg.DupFlag {
		fixHeaderflag |= 1 << publishOffsetDup
	}

	buf := bytes.NewBuffer(nil)
	buf.WriteByte(CtrlTypePUBLISH<<4 | fixHeaderflag)
	buf.Write(encodeLength(remainingLength))
	buf.Write(encodeUint16(uint16(len(msg.Topic))))
	buf.WriteString(msg.Topic)
	if msg.QosLevel != Qos0 {
		buf.Write(encodeUint16(msg.ID))
	}
	buf.Write(msg.Payload)
	if _, err := buf.WriteTo(w); err != nil {
		return err
	}

	return nil
}
