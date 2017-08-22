package mqtt

import (
	"crypto/tls"
	"net/url"
	"time"
)

// Options defines the configuration of mqtt client
type Options struct {
	Servers                 []*url.URL // mqtt servers
	ClientID                string
	Username                string
	Password                string
	CleanSession            bool
	WillEnabled             bool
	WillTopic               string
	WillPayload             []byte
	WillQos                 byte
	WillRetained            bool
	ProtocolVersion         uint
	protocolVersionExplicit bool

	TLSConfig            tls.Config
	KeepAlive            time.Duration
	PingTimeout          time.Duration
	ConnectTimeout       time.Duration // timeout of dailing a server
	MaxReconnectInterval time.Duration
	WriteTimeout         time.Duration
	AutoReconnect        bool
}
