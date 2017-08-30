package packet

import (
	"bytes"
	"io"
)

// TODO: subscribe mutiple
type Subscribe struct {
	ID          uint16
	TopicFilter []string
	QosLevel    []byte
}

func (p *Subscribe) Write(w io.Writer) error {
	var remainingLength int

	if len(p.TopicFilter) != len(p.QosLevel) {
		panic("topic and qos level setting error")
	}

	remainingLength += 2 // Packet Identifier
	for _, top := range p.TopicFilter {
		remainingLength += (2 + len(top) + 1) // Topic + Qos
	}

	buf := bytes.NewBuffer(nil)
	// Bits 3,2,1 and 0 of the fixed header of the SUBSCRIBE Control Packet are reserved and MUST be set to 0,0,1 and 0 respectively.
	// The Server MUST treat any other value as malformed and close the Network Connection [MQTT-3.8.1-1].
	buf.WriteByte(CtrlTypeSUBSCRIBE<<4 | 0x02)
	buf.Write(encodeLength(remainingLength))
	buf.Write(encodeUint16(p.ID))
	for i, top := range p.TopicFilter {
		buf.Write(encodeUint16(uint16(len(top))))
		buf.WriteString(top)
		buf.WriteByte(p.QosLevel[i])
	}

	_, err := buf.WriteTo(w)
	return err
}
