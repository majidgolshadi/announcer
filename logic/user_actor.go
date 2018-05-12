package logic

import "github.com/majidgolshadi/client-announcer/output"

type UserAct struct {
	MessageTemplate string
	Username        string
}

type UserActor struct {
}

func (ua *UserActor) Listen(chanAct <-chan *UserAct, msg chan<- *output.Msg) {
	for rec := range chanAct {
		msg <- &output.Msg{
			Temp: rec.MessageTemplate,
			User: rec.Username,
		}
	}
}
