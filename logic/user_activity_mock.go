package logic

import (
	"strings"
)

type redisMock struct {
	onlineUsers []string
}

func NewUserActivityMock() UserActivity {
	return &redisMock{
		onlineUsers: []string{
			"user1",
			"user2",
			"user3",
			"user4",
			"user5",
			"user6",
		},
	}
}

func (r *redisMock) IsHeOnline(username string) bool {
	return strings.Contains(username, "online")
}

func (r *redisMock) FilterOnlineUsers(usernames []string) []string {
	var result []string

	for _, username := range usernames {
		if r.IsHeOnline(username) {
			result = append(result, username)
		}
	}

	return result
}

func (r *redisMock) Close() {
	return
}
