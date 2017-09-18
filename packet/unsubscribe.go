package packet

import (
	"bytes"
	"io"
)

// TODO: subscribe mutiple
type UnSubscribe struct {
	ID          uint16
	TopicFilter []string
}

func (p *UnSubscribe) Read(w io.Reader) error {
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
