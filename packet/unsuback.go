package packet

import (
	"encoding/binary"
	"io"
)

type UnSubAck struct {
	ID uint16
}

func createUnSubAck(r io.Reader, remainingLen int, fixFlags byte) (interface{}, error) {
	// TODO: move to ReadPacket, read full expect for Publish

	buf := make([]byte, remainingLen)
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, err
	}

	m := &UnSubAck{}
	m.ID = binary.BigEndian.Uint16(buf[:2])
	return m, nil
}
