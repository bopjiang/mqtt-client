package packet

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
)

// TODO: subscribe multiple topics
type UnSubscribe struct {
	FixedHeader
	ID          uint16
	TopicFilter []string
}

func (msg *UnSubscribe) Read(r io.Reader) error {
	buf := make([]byte, msg.RemainingLen)
	if _, err := io.ReadFull(r, buf); err != nil {
		return err
	}

	msg.ID = binary.BigEndian.Uint16(buf[:2])
	buf = buf[2:]
	for {
		flen := binary.BigEndian.Uint16(buf[:2])
		if int(2+flen) > len(buf) {
			return errors.New("extra data in payload")

		}

		topicFilter := string(buf[2 : 2+flen])
		msg.TopicFilter = append(msg.TopicFilter, topicFilter)
		if 2+int(flen) == int(msg.RemainingLen-2) {
			break
		}

		buf = buf[2+flen:]
	}

	return nil
}

func (p *UnSubscribe) Write(w io.Writer) error {
	var remainingLength int

	remainingLength += 2 // Packet Identifier
	for _, top := range p.TopicFilter {
		remainingLength += (2 + len(top))
	}

	buf := bytes.NewBuffer(nil)
	// Bits 3,2,1 and 0 of the fixed header of the SUBSCRIBE Control Packet are reserved and MUST be set to 0,0,1 and 0 respectively.
	// The Server MUST treat any other value as malformed and close the Network Connection [MQTT-3.8.1-1].
	buf.WriteByte(CtrlTypeUNSUBSCRIBE<<4 | 0x02)
	buf.Write(encodeLength(remainingLength))
	buf.Write(encodeUint16(p.ID))
	for _, top := range p.TopicFilter {
		buf.Write(encodeUint16(uint16(len(top))))
		buf.WriteString(top)
	}

	_, err := buf.WriteTo(w)
	return err
}
