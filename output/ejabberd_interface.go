package output

type Ejabberd interface {
	Connect() error
	Send(msg *Msg) error
	Close()
}
