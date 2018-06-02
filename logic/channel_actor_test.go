package logic

import (
	"github.com/majidgolshadi/client-announcer/output"
	"testing"
)

const SoroushChannelId string = "officialsoroushchannel"

func TestGetAllOnlineUsers(t *testing.T) {
	r := NewUserActivityMock()

	ca := &ChannelActor{
		UserActivity: r,
	}

	var userCount int
	reChan, _ := ca.UserActivity.GetAllOnlineUsers()
	for range reChan {
		userCount++
	}

	if userCount != 6 {
		t.Fail()
	}
}

func TestSentToSoroushChannelOnlineUser(t *testing.T) {
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
