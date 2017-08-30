package packet

import (
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
	m.ID = binary.BigEndian.Uint16(buf)

	return m, nil
}
