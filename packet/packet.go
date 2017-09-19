// Package packet is a internal package, and contains MQTT packet encoding and decoding logic.
//
// This package is not a part of API, and supposed to be used .
package packet

import (
	"errors"
	"fmt"
	"io"
	"log"
)

var (
	InvalidPacketLengthErr = errors.New("invalid remaining length")
)

type ControlPacket interface {
	Read(r io.Reader) error
	Write(w io.Writer) error
}

type FixedHeader struct {
	MsgType      byte
	Flag         byte
	RemainingLen uint32 // up to 268,435,455 (256 MB)
}

const (
	maxRemainingLen = 268435455
)

// Command Code
const (
	CtrlTypeReserved1   = byte(0)
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
	CtrlTypeReserved2   = byte(15)
)

func ReadPacket(r io.Reader) (ControlPacket, error) {
	first := make([]byte, 1)
	if _, err := r.Read(first); err != nil {
		return nil, err
	}

	controlType := first[0] >> 4
	fixFlags := first[0] ^ 0x0F
	if controlType == CtrlTypeReserved1 || controlType == CtrlTypeReserved2 {
		log.Printf("read invalid control type, %d", controlType)
		return nil, errors.New("invalid control type")
	}

	remainingLen := decodeLength(r)
	if remainingLen > maxRemainingLen {
		return nil, errors.New("remaining length error")
	}

	h := &FixedHeader{
		MsgType:      controlType,
		Flag:         fixFlags,
		RemainingLen: uint32(remainingLen),
	}

	p := createPacket(h)
	if err := p.Read(r); err != nil {
		return nil, err
	}

	return p, nil
}

func createPacket(h *FixedHeader) ControlPacket {
	switch h.MsgType {
	case CtrlTypeCONNECT:
		return &Connect{FixedHeader: *h}
	case CtrlTypeCONNECTACK:
		return &ConnectAck{FixedHeader: *h}
	case CtrlTypePUBLISH:
		return &Publish{FixedHeader: *h}
	case CtrlTypePUBACK:
		return &PubAck{FixedHeader: *h}
	case CtrlTypeSUBSCRIBE:
		return &Subscribe{FixedHeader: *h}
	case CtrlTypeSUBACK:
		return &SubAck{FixedHeader: *h}
	case CtrlTypeUNSUBSCRIBE:
		return &UnSubscribe{FixedHeader: *h}
	case CtrlTypeUNSUBACK:
		return &UnSubAck{FixedHeader: *h}
	case CtrlTypePINGREQ:
		return &PingReq{FixedHeader: *h}
	case CtrlTypePINGRESP:
		return &PingResp{FixedHeader: *h}
	case CtrlTypeDISCONNECT:
		return &DisConnect{FixedHeader: *h}
	default:
		panic(fmt.Sprintf("invalid msg type, %d", h.MsgType))
	}
}
