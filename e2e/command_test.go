package e2e_test

import (
	"context"
	"log"
	"net/url"
	"runtime"
	"strings"
	"testing"
	"time"

	mqtt "github.com/openim/mqtt-client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type CommandTestSuite struct {
	suite.Suite
	servers     []*url.URL
	cleanServer func()
	c           mqtt.Client
	a           *assert.Assertions
}

// SetupAllSuite has a SetupSuite method, which will run before the
// tests in the suite are run.
func (s *CommandTestSuite) SetupSuite() {
	s.a = assert.New(s.T())
}

func (s *CommandTestSuite) SetupTest() {
	s.servers, s.cleanServer = MustGetMqttServers(s.T())
	opt := mqtt.Options{
		Servers:      s.servers,
		ClientID:     "e2e test client",
		KeepAlive:    time.Second * 1,
		CleanSession: true,
	}

	c := mqtt.NewClient(opt)
	ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
	err := c.Connect(ctx)
	s.a.Nilf(err, "failed to connect, %s", err)
	s.c = c
}

func (s *CommandTestSuite) TearDownTest() {
	// make sure there is any client connection.
	// make sure subscribe storage is cleaned.
	// make sure message storage is cleaned.
	s.c.Disconnect()
	s.cleanServer()
	if goroutineLeaked() {
		s.a.Fail("goroutine leaked")
	}
}

func (s *CommandTestSuite) TearDownSuite() {
}

func (s *CommandTestSuite) TestSubscribe() {
	ctx, _ := context.WithTimeout(context.Background(), time.Second)
	err := s.c.Subscribe(ctx, "test_topic", 0, func(msg mqtt.Message) {
		log.Printf("received msg in test from topic [%s],  %s", msg.Topic(), msg.Payload())
	})

	s.a.Nilf(err, "failed to subsribe, %s", err)
}

func (s *CommandTestSuite) TestPublishQos0() {
	ctx, _ := context.WithTimeout(context.Background(), time.Second)
	err := s.c.Publish(ctx, "test_topic", 0, false, []byte("hello"))
	s.a.Nilf(err, "failed to publish, %s", err)
}

func (s *CommandTestSuite) TestPublishQos1() {
	ctx, _ := context.WithTimeout(context.Background(), time.Second)
	err := s.c.Publish(ctx, "test_topic", 1, false, []byte("hello"))
	s.a.Nilf(err, "failed to publish, %s", err)
}

func (s *CommandTestSuite) TestKeepalive() {
	if testing.Short() {
		return
	}

	time.Sleep(time.Second * 5)
	s.a.True(s.c.IsConnected(), "keepalive failed")
}

func (s *CommandTestSuite) TestUnSubscribe() {
	ctx, _ := context.WithTimeout(context.Background(), time.Second)
	err := s.c.Unsubscribe(ctx, "test_topic", "test_topic2")
	s.a.Nilf(err, "failed to unsubsribe, %s", err)
}

func TestCommandTestSuite(t *testing.T) {
	suite.Run(t, new(CommandTestSuite))
}

func goroutineLeaked() bool {
	buf := make([]byte, 2<<20)
	buf = buf[:runtime.Stack(buf, true)]

	for _, g := range strings.Split(string(buf), "\n\n") {
		sl := strings.SplitN(g, "\n", 2)
		if len(sl) != 2 {
			continue
		}

		// sl[0]: goroutine 10 [running]:
		// sl[1]: github.com/bopjiang/mqtt-client/e2e_test.goroutineLeaked(0xc42004fe58)
		//                /ws/src/github.com/bopjiang/mqtt-client/e2e/command_test.go:51 +0x82
		//        github.com/bopjiang/mqtt-client/e2e_test.(*CommandTestSuite).TearDownTest(0xc420010840)
		//               ....
		stack := strings.TrimSpace(sl[1])
		if strings.Contains(stack, "mqtttest.(*testServer).") ||
			strings.Contains(stack, "mqtt.(*client).") {
			log.Printf("leaked goroutine, %s", stack)
			return true
		}
	}

	return false
}
