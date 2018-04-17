package client_announcer

import (
	"crypto/tls"
	"github.com/soroush-app/xmpp-client/xmpp"
	"time"
)

type ClientSender struct {
	keepConnectionAliveTicker *time.Ticker
	connection                *xmpp.Conn
}

func (n *ClientSender) Connect(address string, username string, password string, domain string, duration time.Duration) (err error) {
	n.connection, err = xmpp.Dial(address, username, domain, "announcer", password, &xmpp.Config{
		SkipTLS:   true,
		TLSConfig: &tls.Config{},
	})

	if err != nil {
		return err
	}

	n.keepConnectionAlive(duration)
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
