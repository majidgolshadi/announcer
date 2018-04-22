package client_announcer

import (
	"fmt"
	"time"

	"github.com/sheenobu/go-xco"
	log "github.com/sirupsen/logrus"
)

type ComponentSender struct {
	Name         string
	Secret       string
	PingInterval int
	Domain       string

	connection                *xco.Component
	keepConnectionAliveTicker *time.Ticker
}

func (cs *ComponentSender) Connect(address string) (err error) {
	cs.connection, err = xco.NewComponent(xco.Options{
		Name:         cs.Name,
		SharedSecret: cs.Secret,
		Address:      address,
	})

	if err != nil {
		return err
	}

	go func() {
		log.Info("connect component ", cs.Name, " to ", address)
		if err := cs.connection.Run(); err != nil {
			log.Fatal("ejabberd component connection error: ", err.Error())
		}
	}()

	cs.keepConnectionAlive(time.Duration(cs.PingInterval) * time.Second)
	return nil
}

func (cs *ComponentSender) keepConnectionAlive(duration time.Duration) {
	cs.keepConnectionAliveTicker = time.NewTicker(duration)
	go func() {
		for range cs.keepConnectionAliveTicker.C {
			cs.connection.Send(fmt.Sprintf("<iq to='%s' type='get'><ping xmlns='urn:xmpp:ping'/></iq>", cs.Domain))
		}
	}()
}

func (cs *ComponentSender) Send(msg string) error {
	_, err := cs.connection.Write([]byte(msg))
	return err
}

func (cs *ComponentSender) Close() {
	log.Info("close component ", cs.Name, " connection")
	cs.keepConnectionAliveTicker.Stop()
	cs.connection.Close()
}
