package logic

import (
	"fmt"
	"time"
)

type redisMock struct {
	onlineUsers map[string]bool
}

func NewRedisMock() UserDataStore {
	return &redisMock{
		onlineUsers: map[string]bool{
			"user1": true,
			"user2": true,
			"user3": true,
			"user4": true,
			"user5": true,
			"user6": true,
		},
	}
}

func (r *redisMock) GetAllOnlineUsers() (<-chan string, error) {
	usersChan := make(chan string)

	go func() {
		for i := 1; i < 7; i++ {
			usersChan <- fmt.Sprintf("user%d", i)
		}
		// if channel closed before read from consumer data will be lost
		time.Sleep(100*time.Millisecond)
		close(usersChan)
	}()

	return usersChan, nil
}

func (r *redisMock) IsHeOnline(username string) bool {
	if r.onlineUsers[username] {
		return true
	}

	return false
}

func (r *redisMock) IsHeOffline(username string) bool {
	return !r.IsHeOnline(username)
}

func (r *redisMock) Close() {
	return
}
