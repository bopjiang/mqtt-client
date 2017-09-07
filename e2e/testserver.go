package e2e

import (
	"fmt"
	"log"
	"net"
	"net/url"
	"testing"
)

// testServer is MQTT test broker
type testServer struct {
	t        *testing.T
	listener net.Listener
}

func mustStartTestServer(t *testing.T) *testServer {
	s := &testServer{t: t}
	s.Start()
	return s
}

func (s *testServer) Start() error {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		s.Errorf("failed to listen, %s", err)
		return nil
	}

	s.listener = listener
	go func() {
		conn, _ := s.listener.Accept()
		log.Printf("new connection, %s -> %s\n", conn.RemoteAddr(), conn.LocalAddr())
		conn.Close()
	}()
	return nil
}

func (s *testServer) Endpoint() *url.URL {
	rawUrl := fmt.Sprintf("tcp://%s", s.listener.Addr().String())
	u, _ := url.Parse(rawUrl)
	return u
}

func (s *testServer) Stop() {

}

func (s *testServer) Errorf(format string, args ...interface{}) {
	s.t.Errorf(format, args)
}
