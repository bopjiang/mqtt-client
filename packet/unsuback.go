package packet

import (
	"bytes"
	"encoding/binary"
	"io"
)

type UnSubAck struct {
	FixedHeader
	ID uint16
}

func (msg *UnSubAck) Read(r io.Reader) error {
	if msg.RemainingLen != 2 {
		return InvalidPacketLengthErr
	}

	buf := make([]byte, msg.RemainingLen)
	if _, err := io.ReadFull(r, buf); err != nil {
		return err
	}

	msg.ID = binary.BigEndian.Uint16(buf[:2])
	return nil
}

func (msg *UnSubAck) Write(w io.Writer) error {
	var remainingLength = 2
	buf := bytes.NewBuffer(nil)
	buf.WriteByte(CtrlTypeUNSUBACK<<4 | 0x02)
	buf.Write(encodeLength(remainingLength))
	buf.Write(encodeUint16(msg.ID))
	_, err := buf.WriteTo(w)
	return err
}
