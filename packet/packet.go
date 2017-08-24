package packet

import (
	"errors"
	"io"
)

type fixedHeader struct {
	MessageType byte
	Flag        byte
}

const (
	ctrlTypeReserved0   = byte(0)
	ctrlTypeCONNECT     = byte(1)
	ctrlTypeCONNECTACK  = byte(2)
	ctrlTypePUBLISH     = byte(3)
	ctrlTypePUBACK      = byte(4)
	ctrlTypePUBREC      = byte(5)
	ctrlTypePUBPUBREL   = byte(6)
	ctrlTypePUBCOMP     = byte(7)
	ctrlTypeSUBSCRIBE   = byte(8)
	ctrlTypeSUBACK      = byte(9)
	ctrlTypeUNSUBSCRIBE = byte(10)
	ctrlTypeUNSUBACK    = byte(11)
	ctrlTypePINGREQ     = byte(12)
	ctrlTypePINGRESP    = byte(13)
	ctrlTypeDISCONNECT  = byte(14)
	ctrlTypeReserved15  = byte(15)
)

type createPacketFunc func(r io.Reader, remainingLen int, fixFlags byte) (interface{}, error)

var createPacketFuncs = []createPacketFunc{
	nil,
	nil,
	createConnectAckPacket,
	nil,
}

// ReadPacket unmarshel a control packet from Reader(normally net.Conn).
// The first parameter returned is a control type specficed packet struct if no error.
func ReadPacket(r io.Reader) (interface{}, error) {
	var firstByte = make([]byte, 1)
	if _, err := io.ReadFull(r, firstByte); err != nil {
		return nil, err
	}

	controlType := firstByte[0] >> 4
	fixFlags := firstByte[0] ^ 0x0F
	remainingLen := decodeLength(r)
	if controlType == ctrlTypeReserved0 || controlType == ctrlTypeReserved15 {
		return nil, errors.New("invalid control type")
	}

	return createPacketFuncs[controlType](r, remainingLen, fixFlags)
}
