package logic

import (
	"github.com/majidgolshadi/client-announcer/output"
	"testing"
	"time"
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

		time.Sleep(50 * time.Millisecond)
		close(out)
	}()

	if err := ca.sentToOnlineUser(SoroushChannelId, "<template to=%s></template>", out); err != nil {
		t.Fail()
	}

	if count != 6 {
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

		time.Sleep(50 * time.Millisecond)
		close(out)
	}()

	if err := ca.sentToOnlineUser("no_online", "<template to=%s></template>", out); err != nil {
		t.Fail()
	}

	if count != 0 {
		t.Fail()
	}
}

func TestSentToChannelWithOnlineUser(t *testing.T) {
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

		time.Sleep(50 * time.Millisecond)
		close(out)
	}()

	if err := ca.sentToOnlineUser("with_online", "<template to=%s></template>", out); err != nil {
		t.Fail()
	}

	if count != 6 {
		println("count is: ", count)
		t.Fail()
	}
}
