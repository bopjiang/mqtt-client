package packet

import (
	"encoding/binary"
	"io"
)

type SubAck struct {
	ID      uint16
	RetCode []byte
}

var SubackReturnCodes = map[uint8]string{
	0:    "Success - Maximum QoS 0",
	1:    "Success - Maximum QoS 1",
	2:    "Success - Maximum QoS 2",
	0x80: "Failure",
}

func createSubAck(r io.Reader, remainingLen int, fixFlags byte) (interface{}, error) {
	// TODO: move to ReadPacket, read full expect for Publish

	buf := make([]byte, remainingLen)
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, err
	}

	m := &SubAck{}
	m.ID = binary.BigEndian.Uint16(buf[:2])
	m.RetCode = buf[2:]
	return m, nil
}
