package client_announcer

type Connection interface {
	Send(msg string) error
	Close()
}
