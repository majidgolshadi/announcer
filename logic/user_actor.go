package logic

import "fmt"

type UserAct struct {
	MessageTemplate string
	Usernames       []string
}

type UserActor struct {
	Domain string
}

func (ua *UserActor) Listen(userAct <-chan *UserAct, msgChan chan<- string, kafkaChan chan<- string) error {
	var message string

	for rec := range userAct {
		for _,user := range rec.Usernames {
			// TODO: Create, validate, marshal and unmarshal message must be move into a struct
			// message creator
			message = fmt.Sprintf(rec.MessageTemplate, fmt.Sprintf("%s@%s", user, ua.Domain))

			kafkaChan <- message
			msgChan <- message
		}
	}

	return nil
}
