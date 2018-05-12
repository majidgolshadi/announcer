package logic

import (
	"fmt"
)

type UserAct struct {
	MessageTemplate string
	Username        string
}

type UserActor struct {
}

func (ua *UserActor) Listen(chanAct <-chan *UserAct, msg chan<- string) {
	for rec := range chanAct {
		msg <- fmt.Sprintf(rec.MessageTemplate, rec.Username)
	}
}
