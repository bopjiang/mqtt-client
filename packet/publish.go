package packet

import (
	"bytes"
	"encoding/binary"
	"io"
)

type Publish struct {
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

func createPublish(r io.Reader, remainingLen int, fixFlags byte) (interface{}, error) {
	m := &Publish{
		RetainFlag: fixFlags|1<<publishOffsetRetain == 1,
		QosLevel:   (fixFlags >> publishOffsetQos) & 0x03,
		DupFlag:    fixFlags|1<<publishOffsetDup == 1,
	}

	buf := make([]byte, remainingLen)
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, err
	}

	topicLen := binary.BigEndian.Uint16(buf[:2])
	m.Topic = string(buf[2 : 2+topicLen])
	if m.QosLevel != Qos0 {
		m.ID = binary.BigEndian.Uint16(buf[2+topicLen : 2+topicLen+2])
	}

	m.Payload = buf[2+topicLen+2:]
	//log.Printf("received in pub %x\n", buf)
	return m, nil
}

func (p *Publish) Write(w io.Writer) error {
	var (
		remainingLength int
		fixHeaderflag   byte
	)

	remainingLength += (2 + len(p.Topic)) // Topic Name
	if p.QosLevel != Qos0 {
		remainingLength += 2 // Packet Identifier
	}
	remainingLength += len(p.Payload)

	if p.RetainFlag {
		fixHeaderflag |= 1 << publishOffsetRetain
	}
	if p.QosLevel > 0 {
		fixHeaderflag |= p.QosLevel << publishOffsetQos
	}
	if p.DupFlag {
		fixHeaderflag |= 1 << publishOffsetDup
	}

	buf := bytes.NewBuffer(nil)
	buf.WriteByte(CtrlTypePUBLISH<<4 | fixHeaderflag)
	buf.Write(encodeLength(remainingLength))
	buf.Write(encodeUint16(uint16(len(p.Topic))))
	buf.WriteString(p.Topic)
	if p.QosLevel != Qos0 {
		buf.Write(encodeUint16(p.ID))
	}
	buf.Write(p.Payload)
	if _, err := buf.WriteTo(w); err != nil {
		return err
	}

	return nil
}
