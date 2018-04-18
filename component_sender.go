package client_announcer

import (
	"github.com/sheenobu/go-xco"
	"time"
	"log"
)

type ComponentSender struct {
	connection                *xco.Component
	keepConnectionAliveTicker *time.Ticker
}

func (cs *ComponentSender) Connect(address string, name string, secret string, duration time.Duration) (err error) {
	cs.connection, err = xco.NewComponent(xco.Options{
		Name:         name,
		SharedSecret: secret,
		Address:      address,
	})

	go func() {
		if err := cs.connection.Run(); err != nil {
			log.Fatal("ejabberd component connection error: ", err.Error())
		}
	}()

	cs.keepConnectionAlive(duration)

	return err
}

func (cs *ComponentSender) keepConnectionAlive(duration time.Duration) {
	cs.keepConnectionAliveTicker = time.NewTicker(duration)
	go func() {
		for range cs.keepConnectionAliveTicker.C {
			cs.connection.Send("<iq to='soroush.ir' type='get'><ping xmlns='urn:xmpp:ping'/></iq>")
		}
	}()
}

func (cs *ComponentSender) Send(msg string) error {
	_, err := cs.connection.Write([]byte(msg))
	return err
}

func (cs *ComponentSender) Close() {
	cs.keepConnectionAliveTicker.Stop()
	cs.connection.Close()
}
