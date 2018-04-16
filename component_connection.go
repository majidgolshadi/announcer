package client_announcer

import (
	"github.com/sheenobu/go-xco"
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

	go func() {
		if err := cc.connection.Run(); err != nil {
			println(err.Error())
		}
	}()

	return err
}

func (cc *ComponentConnection) Send(msg string) error {
	_, err := cc.connection.Write([]byte(msg))
	return err
}

func (cc *ComponentConnection) Close() {
	cc.connection.Close()
}
