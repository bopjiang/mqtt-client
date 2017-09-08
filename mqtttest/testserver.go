package mqtttest

import (
	"fmt"
	"log"
	"net"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/bopjiang/mqtt-client/packet"
)

// testServer is MQTT test broker
type testServer struct {
	t        *testing.T
	listener net.Listener

	exitCh chan struct{}
	wg     sync.WaitGroup
}

func MustStartTestServer(t *testing.T) *testServer {
	s := &testServer{
		t:      t,
		exitCh: make(chan struct{}),
	}
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
	s.wg.Add(1)
	go s.serve()
	log.Printf("test server %s started. ", s.listener.Addr())
	return nil
}

func (s *testServer) Endpoint() *url.URL {
	rawUrl := fmt.Sprintf("tcp://%s", s.listener.Addr().String())
	u, _ := url.Parse(rawUrl)
	return u
}

func (s *testServer) Stop() {
	if s.listener == nil {
		return
	}

	s.listener.Close()
	s.wg.Wait()
	log.Printf("test server %s stopped. ", s.listener.Addr())
}

func (s *testServer) Errorf(format string, args ...interface{}) {
	s.t.Errorf(format, args...)
}

func (s *testServer) serve() {
	defer s.wg.Done()
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			log.Printf("testserver accept err, %s", err)
			return
		}

		s.wg.Add(1)
		go s.handleConn(conn)
	}
}

func (s *testServer) handleConn(conn net.Conn) {
	defer s.wg.Done()
	conn.SetReadDeadline(time.Now().Add(time.Second * 10))
	pkt, err := packet.ReadPacket(conn) // TODO: if serve other protocol other than MQTT, how can we detected it?
	if err != nil {
		s.Errorf("failed to read CONNECT, %s", err)
		return
	}

	msg, ok := pkt.(*packet.Connect)
	if !ok {
		s.Errorf("not Connect msg")
		return
	}

	log.Printf("new mqtt connection, %s -> %s, %+v\n", conn.RemoteAddr(), conn.LocalAddr(), msg)

	ack := &packet.ConnectAck{}
	ack.Write(conn)

	mconn := newMQTTConn(s.t, conn, s.exitCh)
	mconn.SetTimeout(time.Second * time.Duration(msg.Keepalive) * 2)
	mconn.Serve() // might be panic in side?

	// session restore ????
}
