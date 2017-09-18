package packet

import (
	"errors"
	"io"
)

type PingResp struct {
	FixedHeader
}

func (msg *PingResp) Read(r io.Reader) error {
	if msg.RemainingLen != 0 {
		return errors.New("invalid remaining length")
	}

	return nil
}

func (msg *PingResp) Write(w io.Writer) error {
	buf := make([]byte, 2)
	buf[0] = CtrlTypePINGRESP << 4
	// buf[1]: Remaining Length (0)
	_, err := w.Write(buf)
	return err
}
