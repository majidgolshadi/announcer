package logic

type ChannelDataStore interface {
	GetChannelMembers(string) (<-chan string, error)
	Close()
}
