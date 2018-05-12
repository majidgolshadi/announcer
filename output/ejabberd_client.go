package output

import (
	"crypto/tls"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/soroush-app/xmpp-client/xmpp"
	"time"
)

type ejabberdClient struct {
	opt             *EjabberdClientOpt
	conn            *xmpp.Conn
	checkConnTicker *time.Ticker
}

type EjabberdClientOpt struct {
	Host           string
	Username       string
	Password       string
	Domain         string
	ResourcePrefix string
	PingInterval   time.Duration
}

func (opt *EjabberdClientOpt) init() error {
	if len(opt.Host) < 1 {
		opt.Host = "127.0.0.1:5222"
	}

	if opt.Username == "" {
		return errors.New("username does not set")
	}

	if opt.Password == "" {
		return errors.New("password does not set")
	}

	if opt.Domain == "" {
		return errors.New("domain does not set")
	}

	if opt.PingInterval == 0 {
		opt.PingInterval = 5
	}

	opt.PingInterval = opt.PingInterval * time.Second

	return nil
}

func NewEjabberdClient(opt *EjabberdClientOpt) (*ejabberdClient, error) {
	if err := opt.init(); err != nil {
		return nil, err
	}

	return &ejabberdClient{
		opt:             opt,
		checkConnTicker: time.NewTicker(opt.PingInterval),
	}, nil
}

// Every client with specific resource can only has one connection to ejabberd cluster so in order to have
// multiple connection(single connection to every single node) we have to define different resources
// We attached server ip to prefix resource that we get from configuration file
// So for 3 server in single ejabberd cluster we only have 3 connection
func (ec *ejabberdClient) Connect() (err error) {
	name := ec.getClientName()

	log.Info("connect client ", name, " to host ", ec.opt.Host)
	ec.conn, err = xmpp.Dial(ec.opt.Host, ec.opt.Username, ec.opt.Domain,
		ec.getClientResource(), ec.opt.Password, &xmpp.Config{
			SkipTLS:   true,
			TLSConfig: &tls.Config{},
		})

	if err != nil {
		return err
	}

	go ec.keepConnectionAlive()
	return
}

func (ec *ejabberdClient) getClientName() string {
	return fmt.Sprintf("%s@%s", ec.opt.Username, ec.opt.Domain)
}

func (ec *ejabberdClient) getClientResource() string {
	return fmt.Sprintf("%s-%s", ec.opt.ResourcePrefix, ec.opt.Host)
}

func (ec *ejabberdClient) keepConnectionAlive() {
	for range ec.checkConnTicker.C {
		log.Info("ping server ", ec.opt.Host)
		ec.conn.Ping()
	}
}

func (ec *ejabberdClient) Send(msg *Msg) error {
	m := fmt.Sprintf(msg.Temp, fmt.Sprintf("%s@%s", msg.User, ec.opt.Domain))

	if err := ec.conn.SendCustomMsg(m); err != nil {
		log.WithField("error", err.Error()).Error("client send error")
		ec.Close()
		ec.Connect()

		return err
	}

	log.WithField("message", m).Debug("client message sent")
	return nil
}

// We don't have connection close method so we stop ticker
// every inactive connection will be close every "PingInterval" time with ejabberd
func (ec *ejabberdClient) Close() {
	log.Info("close client connection ", ec.getClientName(), " from ", ec.opt.Host)
	ec.checkConnTicker.Stop()
}
