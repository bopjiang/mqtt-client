package packet

import (
	"errors"
	"io"
)

type PingReq struct {
}

func (p *PingReq) Write(w io.Writer) error {
	buf := make([]byte, 2)
	buf[0] = CtrlTypePINGREQ << 4
	// buf[1]: Remaining Length (0)
	_, err := w.Write(buf)
	return err
}

func createPingReq(r io.Reader, remainingLen int, fixFlags byte) (interface{}, error) {
	if remainingLen != 0 {
		return nil, errors.New("invalid remaining length")
	}

	return &PingReq{}, nil
}
