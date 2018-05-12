package output

type Ejabberd interface {
	Connect() error
	Send(msg string) error
	Close()
}
