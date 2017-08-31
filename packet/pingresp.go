package packet

import (
	"errors"
	"io"
)

type PingResp struct {
}

func createPingResp(r io.Reader, remainingLen int, fixFlags byte) (interface{}, error) {
	if remainingLen != 0 {
		return nil, errors.New("invalid remaining length")
	}

	return &PingResp{}, nil
}
