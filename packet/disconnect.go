package packet

import (
	"errors"
	"io"
)

type DisConnect struct {
}

func (msg *DisConnect) Write(w io.Writer) error {
	buf := make([]byte, 2)
	buf[0] = CtrlTypeDISCONNECT << 4
	// buf[1]: Remaining Length (0)
	_, err := w.Write(buf)
	return err
}

func createDisConnect(r io.Reader, remainingLen int, fixFlags byte) (interface{}, error) {
	if remainingLen != 0 {
		return nil, errors.New("invalid remaining length")
	}

	return &DisConnect{}, nil
}
