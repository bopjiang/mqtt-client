package packet

import "io"

type DisConnect struct {
}

func (msg *DisConnect) Write(w io.Writer) error {
	buf := make([]byte, 2)
	buf[0] = CtrlTypeDISCONNECT << 4
	// buf[1]: Remaining Length (0)
	_, err := w.Write(buf)
	return err
}
