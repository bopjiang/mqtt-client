package packet

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
)

// TODO: subscribe mutiple
type Subscribe struct {
	FixedHeader
	ID          uint16
	TopicFilter []string
	QosLevel    []byte
}

func (msg *Subscribe) Read(r io.Reader) error {
	buf := make([]byte, msg.RemainingLen)
	if _, err := io.ReadFull(r, buf); err != nil {
		return err
	}

	msg.ID = binary.BigEndian.Uint16(buf[:2])

	buf = buf[2:]
	for {
		flen := binary.BigEndian.Uint16(buf[:2])
		if int(2+flen+1) > len(buf) {
			return errors.New("extra data in payload")

		}
		topicFilter := string(buf[2 : 2+flen])
		qos := buf[2+flen] & 0x03
		msg.TopicFilter = append(msg.TopicFilter, topicFilter)
		msg.QosLevel = append(msg.QosLevel, qos)
		if 2+int(flen)+1 == int(msg.RemainingLen-2) {
			break
		}

		buf = buf[2+flen+1:]
	}

	return nil

}

func (msg *Subscribe) Write(w io.Writer) error {
	var remainingLength int

	if len(msg.TopicFilter) != len(msg.QosLevel) {
		panic("topic and qos level setting error")
	}

	remainingLength += 2 // Packet Identifier
	for _, top := range msg.TopicFilter {
		remainingLength += (2 + len(top) + 1) // Topic + Qos
	}

	buf := bytes.NewBuffer(nil)
	// Bits 3,2,1 and 0 of the fixed header of the SUBSCRIBE Control Packet are reserved and MUST be set to 0,0,1 and 0 respectively.
	// The Server MUST treat any other value as malformed and close the Network Connection [MQTT-3.8.1-1].
	buf.WriteByte(CtrlTypeSUBSCRIBE<<4 | 0x02)
	buf.Write(encodeLength(remainingLength))
	buf.Write(encodeUint16(msg.ID))
	for i, top := range msg.TopicFilter {
		buf.Write(encodeUint16(uint16(len(top))))
		buf.WriteString(top)
		buf.WriteByte(msg.QosLevel[i])
	}

	_, err := buf.WriteTo(w)
	return err
}
