package logic

import (
	"github.com/majidgolshadi/client-announcer/output"
	"strings"
	"testing"
)

func TestSentToChannelWithNoOnlineUser(t *testing.T) {
	ca := &ChannelActor{
		UserActivity:     NewUserActivityMock(),
		ChannelDataStore: NewMysqlMock(),
	}

	out := make(chan *output.Message)
	count := 0
	go func() {
		for range out {
			count++
		}
	}()

	if err := ca.sentToOnlineUser("no_online", "<template to=%s></template>", out); err != nil {
		t.Log("sent to online user error: ", err.Error())
		t.Fail()
	}

	close(out)

	if count != 0 {
		t.Log("count: ", count)
		t.Fail()
	}
}

func TestSentToChannelWithOnlineUser(t *testing.T) {
	ca := &ChannelActor{
		UserActivity:     NewUserActivityMock(),
		ChannelDataStore: NewMysqlMock(),
	}

	out := make(chan *output.Message)
	msgTemplate := "<template to=%s></template>"
	count := 0
	go func() {
		for range out {
			count++
		}
	}()

	if err := ca.sentToOnlineUser("with_online", msgTemplate, out); err != nil {
		t.Log("sent to online user error: ", err.Error())
		t.Fail()
	}

	close(out)

	if count != 6 {
		t.Log("count is: ", count)
		t.Fail()
	}
}

func TestBulkAskFromUserActivityNoOnlineUser(t *testing.T) {
	ca := &ChannelActor{
		UserActivity:     NewUserActivityMock(),
		ChannelDataStore: NewMysqlMock(),
	}
	ca.buffer = []string{"user1", "user2", "user3"}
	out := make(chan *output.Message)
	count := 0
	go func() {
		for range out {
			count++
		}
	}()

	ca.bulkAskFromUserActivity("<fake to=%s></fake>", out)
	close(out)

	if ca.monitChannelOnlineUserNum != 0 {
		t.Log("count is: ", count)
		t.Fail()
	}
}

func TestBulkAskFromUserActivityOneOnlineUser(t *testing.T) {
	ca := &ChannelActor{
		UserActivity:     NewUserActivityMock(),
		ChannelDataStore: NewMysqlMock(),
	}
	ca.buffer = []string{"user1", "user2_online", "user3"}
	out := make(chan *output.Message)
	count := 0
	go func() {
		for range out {
			count++
		}
	}()

	ca.bulkAskFromUserActivity("<fake to=%s></fake>", out)
	close(out)

	if ca.monitChannelOnlineUserNum != 1 {
		t.Log("count is: ", count)
		t.Fail()
	}
}

func TestAttentionSplitStringBehavior(t *testing.T) {
	apiResponse := `[]`
	trimmed := strings.Trim(apiResponse, "]")
	trimmed = strings.Trim(trimmed, "[")

	if len(strings.Split(trimmed, ",")) != 1 {
		t.Fail()
	}
}
