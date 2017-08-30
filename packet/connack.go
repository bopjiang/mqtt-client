package packet

import "io"

type ConnectAck struct {
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

func createConnectAck(r io.Reader, remainingLen int, fixFlags byte) (interface{}, error) {
	// TODO: move to ReadPacket, read full expect for Publish
	buf := make([]byte, remainingLen)
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, err
	}

	m := &ConnectAck{}
	m.SessionPresent = (buf[0] & 0x01) == 1
	m.ReturnCode = uint8(buf[1])

	return m, nil
}
