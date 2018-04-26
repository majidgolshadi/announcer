package client_announcer

import (
	"testing"
)

func TestGenerateID(t *testing.T) {
	cs := &ComponentSender{}
	id := cs.generateID(5)
	if len(id) != 5 {
		t.Fail()
	}
}
