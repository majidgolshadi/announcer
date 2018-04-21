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

func (n *ClientSender) keepConnectionAlive(duration time.Duration) {
	n.keepConnectionAliveTicker = time.NewTicker(duration)
	go func() {
		for range n.keepConnectionAliveTicker.C {
			n.connection.Ping()
		}
	}()
}

func (n *ClientSender) Send(msg string) error {
	return n.connection.SendCustomMsg(msg)
}

func (n *ClientSender) Close() {
	n.keepConnectionAliveTicker.Stop()
}
