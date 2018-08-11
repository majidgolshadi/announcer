package logic

type UserActivity interface {
	IsHeOnline(string) bool
	FilterOnlineUsers([]string) []string
	Close()
}
