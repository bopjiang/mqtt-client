package mqtt

import (
	"bytes"
	"encoding/binary"
	"io"
)

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

const (
	protocolName    = "MQTT"
	protocolVersion = byte(4)
)

const (
	connectFlagOffsetCleanSession = 1
	connectFlagOffsetWillFlag     = 2
	connectFlagOffsetWillQos      = 3
	connectFlagOffsetWillRetain   = 5
	connectFlagPasswordFlag       = 6
	connectFlagUserNameFlag       = 7
)

type messageConnect struct {
	// head
	//controlPackageType byte // 4-bits: value=1
	//reservedBits       byte // 4-bits: value=0
	remainingLength int
	connectFlags    byte

	// flags
	CleanSessionFlag byte // 0 or 1
	//WillFlag: can be deduced from willTopic
	WillQoS        byte // 0, 1, 2
	WillRetainFlag byte // 0 or 1
	//password/username flag:can be deduced from username/password

	Keepalive uint16

	// payload
	ClientID    string
	WillTopic   string
	WillMessage string
	UserName    string
	Password    string
}

func (msg *messageConnect) Encode() []byte {
	// 2+4 bytes Protocol Name, 1 byte Protocol Level,  1 byte Connect Flags, 2 bytes Keep Alive
	msg.remainingLength = (2 + 4) + 1 + 1 + 2 // =10+payload
	msg.remainingLength += (2 + len(msg.ClientID))
	if len(msg.WillTopic) != 0 {
		msg.remainingLength += (2 + len(msg.WillTopic))
		msg.remainingLength += (2 + len(msg.WillMessage))
		msg.connectFlags |= byte(1) << connectFlagOffsetWillFlag
		msg.connectFlags |= msg.WillQoS << connectFlagOffsetWillQos
		msg.connectFlags |= msg.WillRetainFlag << connectFlagOffsetWillQos
	}

	if len(msg.UserName) != 0 {
		msg.remainingLength += (2 + len(msg.UserName))
		msg.connectFlags |= byte(1) << connectFlagUserNameFlag
	}

	if len(msg.Password) != 0 {
		msg.remainingLength += (2 + len(msg.Password))
		msg.connectFlags |= byte(1) << connectFlagPasswordFlag

	}

	buf := bytes.NewBuffer(nil)
	buf.WriteByte(ctrlTypeCONNECT<<4 | 0)
	buf.Write(encodeLength(msg.remainingLength))

	buf.Write(encodeUint16(uint16(len(protocolName))))
	buf.Write([]byte(protocolName))
	buf.WriteByte(protocolVersion) // mqtt version

	if msg.CleanSessionFlag == 1 {
		msg.connectFlags |= byte(1) << connectFlagOffsetCleanSession
	}

	buf.WriteByte(msg.connectFlags)
	buf.Write(encodeUint16(msg.Keepalive))

	// payload
	buf.Write(encodeUint16(uint16(len(msg.ClientID))))
	buf.WriteString(msg.ClientID)
	if len(msg.WillTopic) != 0 {
		buf.Write(encodeUint16(uint16(len(msg.WillTopic))))
		buf.WriteString(msg.WillTopic)
		buf.Write(encodeUint16(uint16(len(msg.WillMessage))))
		buf.WriteString(msg.WillMessage)
	}

	if len(msg.UserName) != 0 {
		buf.Write(encodeUint16(uint16(len(msg.UserName))))
		buf.WriteString(msg.UserName)
	}

	if len(msg.Password) != 0 {
		buf.Write(encodeUint16(uint16(len(msg.Password))))
		buf.WriteString(msg.Password)
	}

	return buf.Bytes()
}

func decodeUint16(b io.Reader) uint16 {
	num := make([]byte, 2)
	b.Read(num)
	return binary.BigEndian.Uint16(num)
}

func encodeUint16(num uint16) []byte {
	bytes := make([]byte, 2)
	binary.BigEndian.PutUint16(bytes, num)
	return bytes
}

func encodeLength(length int) []byte {
	var encLength []byte
	for {
		digit := byte(length % 128)
		length /= 128
		if length > 0 {
			digit |= 0x80
		}
		encLength = append(encLength, digit)
		if length == 0 {
			break
		}
	}
	return encLength
}

func decodeLength(r io.Reader) int {
	var rLength uint32
	var multiplier uint32
	b := make([]byte, 1)
	for multiplier < 27 { //fix: Infinite '(digit & 128) == 1' will cause the dead loop
		io.ReadFull(r, b)
		digit := b[0]
		rLength |= uint32(digit&127) << multiplier
		if (digit & 128) == 0 {
			break
		}
		multiplier += 7
	}
	return int(rLength)
}
