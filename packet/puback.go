package packet

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
)

type PubAck struct {
	ID uint16
}

func createPubAck(r io.Reader, remainingLen int, fixFlags byte) (interface{}, error) {
	// TODO: move to ReadPacket, read full expect for Publish
	if remainingLen != 2 {
		return nil, errors.New("error remaining length field value")
	}

	buf := make([]byte, remainingLen)
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, err
	}

	m := &PubAck{}
	m.ID = binary.BigEndian.Uint16(buf[:2])

	return m, nil
}

func (p *PubAck) Write(w io.Writer) error {
	buf := bytes.NewBuffer(nil)

	remainingLength := 2

	buf.WriteByte(CtrlTypePUBACK << 4)
	buf.Write(encodeLength(remainingLength))
	buf.Write(encodeUint16(p.ID))
	_, err := buf.WriteTo(w)
	return err
}
