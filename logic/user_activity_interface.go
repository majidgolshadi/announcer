package logic

type UserActivity interface {
	GetAllOnlineUsers() (<-chan string, error)
	IsHeOnline(string) bool
	Close()
}
