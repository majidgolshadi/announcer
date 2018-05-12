package output

import (
	"errors"
	"fmt"
	"github.com/sheenobu/go-xco"
	log "github.com/sirupsen/logrus"
	"math/rand"
	"time"
)

type ejabberdComponent struct {
	opt             *EjabberdComponentOpt
	xcoOpt          xco.Options
	conn            *xco.Component
	checkConnTicker *time.Ticker
}

type EjabberdComponentOpt struct {
	Host         string
	Name         string
	Secret       string
	PingInterval time.Duration
	Domain       string
}

func (opt *EjabberdComponentOpt) init() error {
	if opt.Name == "" {
		return errors.New("name does not set")
	}

	if opt.Secret == "" {
		return errors.New("secret does not set")
	}

	if opt.Domain == "" {
		return errors.New("domain does not set")
	}

	if opt.Host == "" {
		opt.Host = "127.0.0.1:9999"
	}

	if opt.PingInterval == 0 {
		opt.PingInterval = 5
	}

	opt.PingInterval = opt.PingInterval * time.Second

	return nil
}

func NewEjabberdComponent(opt *EjabberdComponentOpt) (*ejabberdComponent, error) {
	if err := opt.init(); err != nil {
		return nil, err
	}

	return &ejabberdComponent{
		opt:             opt,
		checkConnTicker: time.NewTicker(opt.PingInterval),
		xcoOpt: xco.Options{
			Name:         opt.Name,
			Address:      opt.Host,
			SharedSecret: opt.Secret,
		},
	}, nil
}

func (ec *ejabberdComponent) Connect() (err error) {
	if ec.conn, err = xco.NewComponent(ec.xcoOpt); err != nil {
		return err
	}

	go func() {
		log.Info("connect component ", ec.opt.Name, " to ", ec.opt.Host)
		if err := ec.conn.Run(); err != nil {
			log.Error("ejabberd component connection error: ", err.Error())
		}
	}()

	go ec.keepConnectionAlive()
	return nil
}

func (ec *ejabberdComponent) keepConnectionAlive() {
	for range ec.checkConnTicker.C {
		ec.conn.Send(fmt.Sprintf("<iq to='%s' type='get' id='%s'><ping xmlns='urn:xmpp:ping'/></iq>", ec.opt.Domain, generateMsgID(5)))
	}
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")

func generateMsgID(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}

	return string(b)
}

func (ec *ejabberdComponent) Send(msg string) error {
	if _, err := ec.conn.Write([]byte(msg)); err != nil {
		log.WithField("error", err.Error()).Error("component send error")
		ec.Close()
		ec.Connect()

		return err
	}

	return nil
}

func (ec *ejabberdComponent) Close() {
	log.Warn("close component ", ec.opt.Name, " connection from", ec.opt.Host)
	ec.checkConnTicker.Stop()

	if ec.conn != nil {
		ec.conn.Close()
	}
}
