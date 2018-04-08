package client_announcer

import (
	"github.com/soroush-app/xmpp-client/xmpp"
	"crypto/tls"
	"time"
)

type NodeConnection struct {
	keepConnectionAliveTicker *time.Ticker
	connection *xmpp.Conn
}

func (n *NodeConnection) Connect(address string, username string, password string, domain string, duration time.Duration) (err error) {
	n.connection, err = xmpp.Dial(address, username, domain, "announcer", password, &xmpp.Config{
		SkipTLS: true,
		TLSConfig: &tls.Config{},
	})

	if err != nil {
		return err
	}

	n.keepConnectionAlive(duration)
	return
}

func (n *NodeConnection) keepConnectionAlive(duration time.Duration) {
	n.keepConnectionAliveTicker = time.NewTicker(duration)
	go func() {
		for t := range n.keepConnectionAliveTicker.C {
			n.connection.Ping()
			println(t.Second())
		}
	}()
}

func (n *NodeConnection) Send(msg string) error {
	return n.connection.SendCustomMsg(msg)
}

func (n *NodeConnection) Close() {
	n.keepConnectionAliveTicker.Stop()
}