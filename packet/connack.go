package packet

import (
	"bytes"
	"io"
)

type ConnectAck struct {
	FixedHeader
	SessionPresent bool
	ReturnCode     uint8
}

var ConnackReturnCodes = map[uint8]string{
	0: "Connection Accepted",
	1: "Connection Refused: Bad Protocol Version",
	2: "Connection Refused: Client Identifier Rejected",
	3: "Connection Refused: Server Unavailable",
	4: "Connection Refused: Username or Password in unknown format",
	5: "Connection Refused: Not Authorised",
}

func (msg *ConnectAck) Read(r io.Reader) error {
	// TODO: move to ReadPacket, read full expect for Publish
	buf := make([]byte, msg.RemainingLen)
	if _, err := io.ReadFull(r, buf); err != nil {
		return err
	}

	msg.SessionPresent = (buf[0] & 0x01) == 1
	msg.ReturnCode = uint8(buf[1])

	return nil
}

func (msg *ConnectAck) Write(w io.Writer) error {
	buf := bytes.NewBuffer(nil)

	remainingLength := 2

	buf.WriteByte(CtrlTypeCONNECTACK << 4)
	buf.Write(encodeLength(remainingLength))

	var flag byte
	if msg.SessionPresent {
		flag |= 0x01
	}
	buf.WriteByte(flag)
	buf.WriteByte(msg.ReturnCode)
	_, err := buf.WriteTo(w)
	return err
}
