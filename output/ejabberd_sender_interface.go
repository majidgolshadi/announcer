package output

type EjabberdSender interface {
	Connect() error
	Send(msg string) error
	Close()
}
