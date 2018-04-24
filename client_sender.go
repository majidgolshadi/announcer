package client_announcer

import (
	"crypto/tls"
	"time"

	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/soroush-app/xmpp-client/xmpp"
)

type ClientSender struct {
	Username     string
	Password     string
	Domain       string
	Resource     string
	PingInterval int

	clientName                string
	connectedToHost           string
	keepConnectionAliveTicker *time.Ticker
	connection                *xmpp.Conn
}

// Every client with specific resource can only has one connection to ejabberd cluster so in order to have
// multiple connection(single connection to every single node) we have to define different resources
// We attached server ip to prefix resource that we get from configuration file
// So for 3 server in single ejabberd cluster we only have 3 connection
func (cs *ClientSender) Connect(host string) (err error) {
	cs.clientName = fmt.Sprintf("%s@%s", cs.Username, cs.Domain)
	cs.connectedToHost = host

	log.Info("connect client ", cs.clientName, " to ", host)
	cs.connection, err = xmpp.Dial(host, cs.Username, cs.Domain, cs.clientResource(host), cs.Password, &xmpp.Config{
		SkipTLS:   true,
		TLSConfig: &tls.Config{},
	})

	if err != nil {
		return err
	}

	cs.keepConnectionAlive(time.Duration(cs.PingInterval) * time.Second)
	return
}

func (cs *ClientSender) clientResource(uniqueKey string) string {
	return fmt.Sprintf("%s-%s", cs.Resource, uniqueKey)
}

func (cs *ClientSender) keepConnectionAlive(duration time.Duration) {
	cs.keepConnectionAliveTicker = time.NewTicker(duration)
	go func() {
		for range cs.keepConnectionAliveTicker.C {
			log.Info(cs.clientName, " ping server ", cs.connectedToHost)
			cs.connection.Ping()
		}
	}()
}

func (cs *ClientSender) Send(msg string) error {
	return cs.connection.SendCustomMsg(msg)
}

func (cs *ClientSender) Close() {
	log.Info("close client ", cs.clientName, " connection")
	cs.keepConnectionAliveTicker.Stop()
}
