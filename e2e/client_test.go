// Package contains end to end client test cases
package e2e_test

import (
	"context"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	mqtt "github.com/bopjiang/mqtt-client"
)

// test server injected using env.
// eg: MQTT_TEST_SERVERS = "tcp://127.0.0.1:1083,tcp://127.0.0.1:1084"
const EnvMqttTestServers = "MQTT_TEST_SERVERS"

var Servers []string

func MustEnv(t *testing.T, key string) (value string) {
	if value = os.Getenv(key); value == "" {
		t.Errorf("ENV %q is not set.", key)
	}
	return value
}

func MustGetMqttServers(t *testing.T) (servers []*url.URL) {
	env := MustEnv(t, EnvMqttTestServers)
	ss := strings.Split(env, ",")
	for _, s := range ss {
		url, err := url.Parse(s)
		if err != nil {
			t.Errorf("failed to parse server url, %s, %s", s, err)
			return
		}
		servers = append(servers, url)
	}

	if len(servers) == 0 {
		t.Errorf("no server configured")
	}

	return
}

func TestConnClient(t *testing.T) {
	servers := MustGetMqttServers(t)
	opt := mqtt.Options{
		Servers:   servers,
		ClientID:  "e2e test client",
		KeepAlive: time.Second * 10,
	}

	c := mqtt.NewClient(opt)
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	if err := c.Connect(ctx); err != nil {
		t.Errorf("failed to connect, %s", err)
		return
	}

}
