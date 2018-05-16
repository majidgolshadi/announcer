package logic

import (
	"github.com/majidgolshadi/client-announcer/output"
	"testing"
)

func TestGetAllOnlineUsers(t *testing.T) {
	r := NewRedisMock()

	ca := &ChannelActor{
		UserDataStore: r,
	}

	var userCount int
	reChan, _ := ca.UserDataStore.GetAllOnlineUsers()
	for range reChan {
		userCount++
	}

	if userCount != 6 {
		t.Fail()
	}
}

func TestSentToSoroushChannelOnlineUser(t *testing.T) {
	ca := &ChannelActor{
		UserDataStore:    NewRedisMock(),
		ChannelDataStore: NewMysqlMock(),
	}

	out := make(chan *output.Msg)
	count := 0
	go func() {
		for range out {
			count++
		}
	}()

	if err := ca.sentToOnlineUser(SoroushChannelId, "<template to=%s></template>", out); err != nil {
		t.Log("sent to online user error: ", err.Error())
		t.Fail()
	}

	close(out)

	if count != 6 {
		t.Log("count: ", count)
		t.Fail()
	}
}

func TestSentToChannelWithNoOnlineUser(t *testing.T) {
	ca := &ChannelActor{
		UserDataStore:    NewRedisMock(),
		ChannelDataStore: NewMysqlMock(),
	}

	out := make(chan *output.Msg)
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
		UserDataStore:    NewRedisMock(),
		ChannelDataStore: NewMysqlMock(),
	}

	out := make(chan *output.Msg)
	msgTemplate := "<template to=%s></template>"
	count := 0
	go func() {
		for msg := range out {
			if msg.Temp != msgTemplate {
				t.Log("invalid template")
				t.Fail()
			}

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
