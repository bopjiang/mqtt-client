package packet

import "io"

type PingReq struct {
}

func (p *PingReq) Write(w io.Writer) error {
	buf := make([]byte, 2)
	buf[0] = CtrlTypePINGREQ << 4
	_, err := w.Write(buf)
	return err
}
