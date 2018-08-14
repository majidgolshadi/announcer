package logic

import (
	"testing"
	"fmt"
)

func TestIsHeOnline(t *testing.T) {
	opt := &UserActivityRestApiOpt{
		RestApiBaseUrl: "172.16.213.65:8888",
	}

	ra, er := NewUserActivityRestApi(opt)
	if er != nil {
		println(er.Error())
	}

	if ra.IsHeOnline("majid") {
		t.Fail()
	}
}

func TestWhichOneIsOnline(t *testing.T) {
	opt := &UserActivityRestApiOpt{
		RestApiBaseUrl: "172.17.32.200:8888",
	}

	ra, er := NewUserActivityRestApi(opt)
	if er != nil {
		println(er.Error())
	}

	resp := ra.FilterOnlineUsers([]string{"up29cndvm", "javad"})

	if len(resp) != 1 {
		t.Fail()
	}

	res := fmt.Sprintf("%s@%s", resp[0], "domain")
	println(res)
}

func TestFilterOnOnlineUser(t *testing.T) {
	opt := &UserActivityRestApiOpt{
		RestApiBaseUrl: "172.16.213.65:8888",
	}

	ra, er := NewUserActivityRestApi(opt)
	if er != nil {
		println(er.Error())
	}

	ra.FilterOnlineUsers([]string{"up29cndvm", "javad"})
}