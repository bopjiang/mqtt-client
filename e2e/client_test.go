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

func MustConnectServer(t *testing.T, clientOpt *mqtt.Options) (c mqtt.Client, cleanFn CleanFn) {
	servers, servCleanfn := MustGetMqttServers(t)

	opt := mqtt.Options{
		Servers:      servers,
		ClientID:     "e2e test client",
		KeepAlive:    time.Second * 5,
		CleanSession: true,
	}

	if clientOpt != nil {
		if clientOpt.KeepAlive != 0 {
			opt.KeepAlive = clientOpt.KeepAlive
		}
	}

	c = mqtt.NewClient(opt)
	cleanFn = func() {
		if err := c.Disconnect(); err != nil {
			log.Printf("client disconnect error, %s", err)
		}

		servCleanfn()
	}

	ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
	if err := c.Connect(ctx); err != nil {
		t.Errorf("failed to connect, %s", err)
		return
	}

	return
}

func TestSubscribe(t *testing.T) {
	c, cleanFn := MustConnectServer(t, nil)
	defer cleanFn()

	ctx, _ := context.WithTimeout(context.Background(), time.Second)
	err := c.Subscribe(ctx, "test_topic", 0, func(msg mqtt.Message) {
		t.Logf("received msg in test from topic [%s],  %s", msg.Topic(), msg.Payload())
	})

	if err != nil {
		t.Errorf("failed to subsribe, %s", err)
	}
}

func TestPublish(t *testing.T) {

}

func TestKeepalive(t *testing.T) {
	if testing.Short() {
		return
	}

	keepAliveTime := time.Second * 1
	c, cleanFn := MustConnectServer(t, &mqtt.Options{KeepAlive: keepAliveTime})
	defer cleanFn()

	time.Sleep(keepAliveTime * 5)
	if !c.IsConnected() {
		t.Errorf("keepalive failed")
	}
}
