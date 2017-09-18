package packet

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
)

type PubAck struct {
	FixedHeader
	ID uint16
}

func (msg *PubAck) Read(r io.Reader) error {
	// TODO: move to ReadPacket, read full expect for Publish
	if msg.RemainingLen != 2 {
		return errors.New("error remaining length field value")
	}

	buf := make([]byte, msg.RemainingLen)
	if _, err := io.ReadFull(r, buf); err != nil {
		return err
	}

	msg.ID = binary.BigEndian.Uint16(buf[:2])

	return nil
}

func (msg *PubAck) Write(w io.Writer) error {
	buf := bytes.NewBuffer(nil)

	remainingLength := 2

	buf.WriteByte(CtrlTypePUBACK << 4)
	buf.Write(encodeLength(remainingLength))
	buf.Write(encodeUint16(msg.ID))
	_, err := buf.WriteTo(w)
	return err
}
