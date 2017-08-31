package packet

import (
	"errors"
	"io"
	"log"
)

type PacketWriter interface {
	Write(w io.Writer) error
}

type fixedHeader struct {
	MessageType byte
	Flag        byte
}

const (
	CtrlTypeReserved0   = byte(0)
	CtrlTypeCONNECT     = byte(1)
	CtrlTypeCONNECTACK  = byte(2)
	CtrlTypePUBLISH     = byte(3)
	CtrlTypePUBACK      = byte(4)
	CtrlTypePUBREC      = byte(5)
	CtrlTypePUBPUBREL   = byte(6)
	CtrlTypePUBCOMP     = byte(7)
	CtrlTypeSUBSCRIBE   = byte(8)
	CtrlTypeSUBACK      = byte(9)
	CtrlTypeUNSUBSCRIBE = byte(10)
	CtrlTypeUNSUBACK    = byte(11)
	CtrlTypePINGREQ     = byte(12)
	CtrlTypePINGRESP    = byte(13)
	CtrlTypeDISCONNECT  = byte(14)
	CtrlTypeReserved15  = byte(15)
)

type createPacketFunc func(r io.Reader, remainingLen int, fixFlags byte) (interface{}, error)

var createPacketFuncs = []createPacketFunc{
	nil,
	nil,              // 1
	createConnectAck, // 2
	createPublish,    // 3
	createPubAck,     // 4
	nil,              // 5
	nil,              // 6
	nil,              // 7
	nil,              // 8
	createSubAck,     // 9
	nil,              //10
	nil,              //11
	nil,              //12
	createPingResp,   //13

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
	if controlType == CtrlTypeReserved0 || controlType == CtrlTypeReserved15 {
		log.Printf("read invalid control type, %d", controlType)
		return nil, errors.New("invalid control type")
	}

	remainingLen := decodeLength(r)
	//log.Printf("read next package, type=%d, len=%d", controlType, remainingLen)
	return createPacketFuncs[controlType](r, remainingLen, fixFlags)
}
