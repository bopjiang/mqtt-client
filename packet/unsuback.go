package packet

import (
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

func (msg *UnSubAck) Write(r io.Writer) error {
	return nil
}
