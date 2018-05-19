package logic

type UserDataStore interface {
	GetAllOnlineUsers() (<-chan string, error)
	IsHeOnline(string) bool
	Close()
}
