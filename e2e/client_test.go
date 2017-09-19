// Package contains end to end client test cases
package e2e_test

import (
	"context"
	"log"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	mqtt "github.com/openim/mqtt-client"
	"github.com/openim/mqtt-client/mqtttest"
)

// all test use intenal test server or external server if configed.
// external server has higher priority if available

// test server injected using env.
// eg: MQTT_TEST_SERVERS = "tcp://127.0.0.1:1083,tcp://127.0.0.1:1084"
const EnvMqttTestServers = "MQTT_TEST_SERVERS"

func MustGetMqttServers(t *testing.T) (servers []*url.URL, cleanFn func()) {
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

func MustConnectServer(t *testing.T, clientOpt *mqtt.Options) (c mqtt.Client, cleanFn func()) {
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
