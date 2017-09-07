// Package contains end to end client test cases
package e2e

import (
	"context"
	"log"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	mqtt "github.com/bopjiang/mqtt-client"
	"github.com/bopjiang/mqtt-client/mqtttest"
)

// all test use intenal test server or external server if configed.
// external server has higher priority if available

// test server injected using env.
// eg: MQTT_TEST_SERVERS = "tcp://127.0.0.1:1083,tcp://127.0.0.1:1084"
const EnvMqttTestServers = "MQTT_TEST_SERVERS"

type CleanFn func()

func MustGetMqttServers(t *testing.T) (servers []*url.URL, cleanFn CleanFn) {
	env := os.Getenv(EnvMqttTestServers)
	ss := strings.Split(env, ",")
	for _, s := range ss {
		s = strings.TrimSpace(s)
		if len(s) == 0 {
			continue
		}
		url, err := url.Parse(s)
		if err != nil {
			t.Errorf("failed to parse server url, %s, %s", s, err)
			return
		}
		servers = append(servers, url)
	}

	if len(servers) == 0 {
		s := mqtttest.MustStartTestServer(t)
		cleanFn = func() { s.Stop() }
		servers = append(servers, s.Endpoint())
		log.Printf("using internal mqtt server, %s\n", servers)
	} else {
		log.Printf("using external mqtt server, %v\n", servers)
	}

	return
}

func TestConnClient(t *testing.T) {
	servers, cleanFn := MustGetMqttServers(t)
	if cleanFn != nil {
		defer cleanFn()
	}

	// TODO: if internal server started, should be stopped when test finished.
	opt := mqtt.Options{
		Servers:      servers,
		ClientID:     "e2e test client",
		KeepAlive:    time.Second * 5,
		CleanSession: true,
	}

	c := mqtt.NewClient(opt)
	defer c.Disconnect()
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	if err := c.Connect(ctx); err != nil {
		t.Errorf("failed to connect, %s", err)
		return
	}
}

func TestConnClient2(t *testing.T) {
	servers, cleanFn := MustGetMqttServers(t)
	if cleanFn != nil {
		defer cleanFn()
	}

	// TODO: if internal server started, should be stopped when test finished.
	opt := mqtt.Options{
		Servers:      servers,
		ClientID:     "e2e test client",
		KeepAlive:    time.Second * 5,
		CleanSession: false,
	}

	c := mqtt.NewClient(opt)
	defer c.Disconnect()
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	if err := c.Connect(ctx); err != nil {
		t.Errorf("failed to connect, %s", err)
		return
	}

	if err := c.Subscribe(ctx, "test/jj2", 1, func(msg mqtt.Message) {
		t.Logf("received msg in test from topic [%s],  %s", msg.Topic(), msg.Payload())
	}); err != nil {
		t.Errorf("failed to subcribe, %s", err)
		return
	}

	data := []byte("test123")
	if err := c.Publish(ctx, "test/jj", 1, false, data); err != nil {
		t.Errorf("failed to publish, %s", err)
		return
	}

	time.Sleep(15 * time.Second)
}

// TODO: read very big payload > 100M
