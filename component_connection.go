package client_announcer

import (
	"github.com/sheenobu/go-xco"
	"time"
)

type ComponentConnection struct {
	connection *xco.Component
}

func (cc *ComponentConnection) Connect(address string, name string, secret string) (err error) {
	cc.connection, err = xco.NewComponent(xco.Options{
		Name:         name,
		SharedSecret: secret,
		Address:      address,
	})

	return err
}

func (cc *ComponentConnection) keepConnectionAlive(duration time.Duration) {
	go func() {
		if err := cc.connection.Run(); err != nil {
			println(err.Error())
		}
	}()
}

func (cc *ComponentConnection) Send(msg string) error {
	return cc.connection.Send(msg)
}

func (cc *ComponentConnection) Close() {
	cc.connection.Close()
}