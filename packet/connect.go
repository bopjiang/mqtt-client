package packet

import (
	"bytes"
	"io"
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

type Connect struct {
	// flags
	CleanSessionFlag bool // 0 or 1
	//WillFlag: can be deduced from willTopic

	WillQoS        byte // 0, 1, 2
	WillRetainFlag bool // 0 or 1
	//password/username flag:can be deduced from username/password

	Keepalive uint16

	// payload
	ClientID    string
	WillTopic   string
	WillMessage []byte
	UserName    string
	Password    string
}

func (msg *Connect) Write(w io.Writer) error {
	// 2+4 bytes Protocol Name, 1 byte Protocol Level,  1 byte Connect Flags, 2 bytes Keep Alive
	var (
		connectFlags    byte
		remainingLength int
	)

	remainingLength = (2 + 4) + 1 + 1 + 2 // =10+payload

	remainingLength += (2 + len(msg.ClientID))
	if len(msg.WillTopic) != 0 {
		remainingLength += (2 + len(msg.WillTopic))
		remainingLength += (2 + len(msg.WillMessage))
		connectFlags |= byte(1) << connectFlagOffsetWillFlag
		connectFlags |= msg.WillQoS << connectFlagOffsetWillQos
		if msg.WillRetainFlag {
			connectFlags |= 1 << connectFlagOffsetWillQos
		}
	}

	if len(msg.UserName) != 0 {
		remainingLength += (2 + len(msg.UserName))
		connectFlags |= byte(1) << connectFlagUserNameFlag
	}

	if len(msg.Password) != 0 {
		remainingLength += (2 + len(msg.Password))
		connectFlags |= byte(1) << connectFlagPasswordFlag

	}

	buf := bytes.NewBuffer(nil)
	buf.WriteByte(CtrlTypeCONNECT << 4)
	buf.Write(encodeLength(remainingLength))

	buf.Write(encodeUint16(uint16(len(protocolName))))
	buf.Write([]byte(protocolName))
	buf.WriteByte(protocolVersion) // mqtt version

	if msg.CleanSessionFlag {
		connectFlags |= byte(1) << connectFlagOffsetCleanSession
	}

	buf.WriteByte(connectFlags)
	buf.Write(encodeUint16(msg.Keepalive))

	// payload
	buf.Write(encodeUint16(uint16(len(msg.ClientID))))
	buf.WriteString(msg.ClientID)
	if len(msg.WillTopic) != 0 {
		buf.Write(encodeUint16(uint16(len(msg.WillTopic))))
		buf.WriteString(msg.WillTopic)
		buf.Write(encodeUint16(uint16(len(msg.WillMessage))))
		buf.Write(msg.WillMessage)
	}

	if len(msg.UserName) != 0 {
		buf.Write(encodeUint16(uint16(len(msg.UserName))))
		buf.WriteString(msg.UserName)
	}

	if len(msg.Password) != 0 {
		buf.Write(encodeUint16(uint16(len(msg.Password))))
		buf.WriteString(msg.Password)
	}

	_, err := buf.WriteTo(w)
	return err
}
