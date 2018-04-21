package client_announcer

type Sender interface {
	Connect(host string) error
	Send(msg string) error
	Close()
}
