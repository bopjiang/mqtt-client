package packet

import (
	"errors"
	"io"
)

type PingReq struct {
	FixedHeader
}

func (msg *PingReq) Read(r io.Reader) error {
	if msg.RemainingLen != 0 {
		return errors.New("invalid remaining length")
	}

	return nil
}

func (msg *PingReq) Write(w io.Writer) error {
	buf := make([]byte, 2)
	buf[0] = CtrlTypePINGREQ << 4
	// buf[1]: Remaining Length (0)
	_, err := w.Write(buf)
	return err
}
