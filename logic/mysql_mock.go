package logic

import "fmt"

type mysqlMock struct {
}

func NewMysqlMock() ChannelDataStore {
	return &mysqlMock{}
}

// have_online channel has 20 user and 6 online
// no_online channel has 14 user and 0 online
func (ms *mysqlMock) GetChannelMembers(channelID string) (<-chan string, error) {
	membersChan := make(chan string)

	go func() {
		i := 1
		if channelID == "no_online" {
			i = 7
		}

		for ; i < 21; i++ {
			membersChan <- fmt.Sprintf("user%d", i)
		}

		close(membersChan)
	}()

	return membersChan, nil
}

func (ms *mysqlMock) Close() {
	return
}
