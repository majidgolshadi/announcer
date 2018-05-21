package output

type EjabberdSender interface {
	Connect() error
	Send(msg *Msg) error
	Close()
}
