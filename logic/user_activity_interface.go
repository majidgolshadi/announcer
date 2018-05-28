package logic

type UserActivity interface {
	GetAllOnlineUsers() (<-chan string, error)
	IsHeOnline(string) bool
	WhichOneIsOnline([]string) []interface{}
	Close()
}
