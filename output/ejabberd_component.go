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
	opt                *EjabberdComponentOpt
	xcoOpt             xco.Options
	conn               *xco.Component
	checkReqConnTicker *time.Ticker
	pingConnTicker     *time.Ticker
	connState          bool
}

type EjabberdComponentOpt struct {
	Host                 string
	Name                 string
	Secret               string
	Domain               string
	PingInterval         time.Duration
	ConnReqCheckInterval time.Duration
	MaxConnCheckRetry    int
}

const PingIq = "<iq to='%s' type='get' id='%s'><ping xmlns='urn:xmpp:ping'/></iq>"

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

	if opt.MaxConnCheckRetry == 0 {
		opt.PingInterval = 10000
	}

	if opt.ConnReqCheckInterval == 0 {
		opt.ConnReqCheckInterval = 10
	}

	opt.PingInterval = opt.PingInterval * time.Second
	opt.ConnReqCheckInterval = opt.ConnReqCheckInterval * time.Millisecond

	return nil
}

func NewEjabberdComponent(opt *EjabberdComponentOpt) (*ejabberdComponent, error) {
	if err := opt.init(); err != nil {
		return nil, err
	}

	ec := &ejabberdComponent{
		opt:       opt,
		connState: false,
		xcoOpt: xco.Options{
			Name:         opt.Name,
			Address:      opt.Host,
			SharedSecret: opt.Secret,
		},
	}

	return ec, nil
}

func (ec *ejabberdComponent) Connect() (err error) {
	if ec.conn, err = xco.NewComponent(ec.xcoOpt); err != nil {
		return err
	}

	failedConnFlag := false
	go func() {
		log.Info("connect component ", ec.opt.Name, " to ", ec.opt.Host)
		if err := ec.conn.Run(); err != nil {
			failedConnFlag = true
			log.Error("ejabberd component connection error: ", err.Error())
		}
	}()

	retry := 0
	ec.checkReqConnTicker = time.NewTicker(ec.opt.ConnReqCheckInterval)
	for range ec.checkReqConnTicker.C {
		retry++
		if failedConnFlag || retry > ec.opt.MaxConnCheckRetry {
			ec.checkReqConnTicker.Stop()
			break
		}

		if ec.ping() == nil {
			ec.checkReqConnTicker.Stop()
			ec.connState = true
			go ec.keepConnectionAlive()
			break
		}
	}

	return nil
}

func (ec *ejabberdComponent) keepConnectionAlive() {
	ec.pingConnTicker = time.NewTicker(ec.opt.PingInterval)

	for range ec.pingConnTicker.C {
		ec.ping()
	}
}

func (ec *ejabberdComponent) ping() error {
	return ec.conn.Send(fmt.Sprintf(PingIq, ec.opt.Domain, generateMsgID(5)))
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
	if ec.conn == nil {
		ec.Connect()
		return errors.New("component connection does not established")
	}

	if !ec.connState {
		return errors.New("component connecting error")
	}

	if _, err := ec.conn.Write([]byte(msg)); err != nil {
		log.WithField("error", err.Error()).Error("component send error")
		ec.Close()
		ec.Connect()

		return err
	}

	log.WithField("message", msg).Debug("component message sent")
	return nil
}

func (ec *ejabberdComponent) Close() {
	log.Warn("close component ", ec.opt.Name, " connection from", ec.opt.Host)
	ec.pingConnTicker.Stop()

	if ec.conn != nil {
		ec.conn.Close()
	}

	ec.connState = false
}
