package client_announcer

import (
	"testing"
	"time"
)

func TestInitRedisOption(t *testing.T) {
	add := "800.800.800.800:2332"
	pass := "password"
	db := 0

	r := &redisDs{
		Address:       add,
		Password:      pass,
		Database:      db,
		CheckInterval: 2,
	}

	r.connect()
	opt := r.connection().Options()
	if opt.Addr != add || opt.Password != pass || opt.DB != db {
		t.Fail()
	}
}

func TestChangeRedisConnectionStatus(t *testing.T) {
	r := &redisDs{
		Address:       "800.800.800.800:3333",
		Password:      "pass",
		Database:      0,
		CheckInterval: 1,
	}

	r.connect()
	time.Sleep(2 * time.Second)
	if r.connectionStatus() != false {
		t.Fail()
	}
}
