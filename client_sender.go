package client_announcer

import (
	"crypto/tls"
	"github.com/soroush-app/xmpp-client/xmpp"
	"time"
)

type ClientSender struct {
	Username     string
	Password     string
	Domain       string
	PingInterval int

	keepConnectionAliveTicker *time.Ticker
	connection                *xmpp.Conn
}

func (cs *ClientSender) Connect(host string) (err error) {
	cs.connection, err = xmpp.Dial(host, cs.Username, cs.Domain, "announcer", cs.Password, &xmpp.Config{
		SkipTLS:   true,
		TLSConfig: &tls.Config{},
	})

	if err != nil {
		return err
	}

	cs.keepConnectionAlive(time.Duration(cs.PingInterval) * time.Second)
	return
}

func (cs *ClientSender) keepConnectionAlive(duration time.Duration) {
	cs.keepConnectionAliveTicker = time.NewTicker(duration)
	go func() {
		for range cs.keepConnectionAliveTicker.C {
			cs.connection.Ping()
		}
	}()
}

func (cs *ClientSender) Send(msg string) error {
	return cs.connection.SendCustomMsg(msg)
}

func (cs *ClientSender) Close() {
	cs.keepConnectionAliveTicker.Stop()
}
