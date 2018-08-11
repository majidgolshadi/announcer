package logic

import "testing"

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
		RestApiBaseUrl: "172.16.213.65:8888",
	}

	ra, er := NewUserActivityRestApi(opt)
	if er != nil {
		println(er.Error())
	}

	ra.FilterOnlineUsers([]string{"up29cndvm", "javad"})
}
