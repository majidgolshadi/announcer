package client_announcer

type Sender interface {
	Send(msg string) error
	Close()
}
