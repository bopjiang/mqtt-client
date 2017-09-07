package packet

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
)

const (
	protocolName  = "MQTT"
	protocolLevel = byte(4)
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

func createConnect(r io.Reader, remainingLen int, fixFlags byte) (interface{}, error) {
	buf := make([]byte, remainingLen)
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, err
	}

	msg := &Connect{}

	// === The variable header ====
	// for the CONNECT Packet consists of four fields in the following order:
	// Protocol Name, Protocol Level, Connect Flags, and Keep Alive

	// Protocol Name
	if binary.BigEndian.Uint16(buf[:2]) != 4 {
		return nil, errors.New("invalid protocol string len")
	}

	if string(buf[2:2+4]) != protocolName {
		return nil, errors.New("invalid protocol name")
	}

	// Protocol Level
	v := uint8(buf[2+4])
	if v != protocolLevel {
		return nil, errors.New("invalid protocol level")
	}

	// Connect Flags
	connectFlags := buf[2+4+1]
	msg.CleanSessionFlag = (connectFlags >> connectFlagOffsetCleanSession & 0x01) == 1
	willFlag := (connectFlags >> connectFlagOffsetWillFlag & 0x01) == 1
	msg.WillQoS = connectFlags >> connectFlagOffsetWillQos & 0x01
	msg.WillRetainFlag = (connectFlags >> connectFlagOffsetWillRetain & 0x01) == 1
	userNameFlag := (connectFlags >> connectFlagUserNameFlag & 0x01) == 1
	passwordFlag := (connectFlags >> connectFlagPasswordFlag & 0x01) == 1

	// Keep Alive
	buf = buf[2+4+1+1:]
	msg.Keepalive = binary.BigEndian.Uint16(buf[:2])

	// =====Payload======
	// These fields, if present, MUST appear in the order :
	// Client Identifier, Will Topic, Will Message, User Name, Password [MQTT-3.1.3-1].

	payload := buf[2:]
	// Client Identifier
	clientIDLen := binary.BigEndian.Uint16(payload[:2]) //TODO: zero-byte ClientId
	msg.ClientID = string(payload[2 : 2+clientIDLen])
	payload = payload[2+clientIDLen:]
	// Will Topic
	// Will Message
	if willFlag {
		willTopicLen := binary.BigEndian.Uint16(payload[:2])
		msg.WillTopic = string(payload[2 : 2+willTopicLen])
		payload = payload[2+willTopicLen:]

		willMessageLen := binary.BigEndian.Uint16(payload[:2])
		msg.WillMessage = payload[2 : 2+willMessageLen]
		payload = payload[2+willMessageLen:]
	}

	// User Name
	if userNameFlag {
		userNameLen := binary.BigEndian.Uint16(payload[:2])
		msg.UserName = string(payload[2 : 2+userNameLen])
		payload = payload[2+userNameLen:]
	}

	// Password
	if passwordFlag {
		passwordLen := binary.BigEndian.Uint16(payload[:2])
		msg.Password = string(payload[2 : 2+passwordLen])
	}
	return msg, nil
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
	buf.WriteByte(protocolLevel) // mqtt version

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
