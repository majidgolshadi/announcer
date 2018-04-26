package client_announcer

import "testing"

func TestMultipleSliceDifference(t *testing.T) {
	slice1 := []string{"a", "b"}
	slice2 := []string{"a"}

	diff := difference(slice1, slice2)
	if diff != "b" {
		t.Fail()
	}

	diff = difference(slice2, slice1)
	if diff != "b" {
		t.Fail()
	}
}

func TestZookeeperFindEvent(t *testing.T) {
	zk := &zookeeper{}
	snapshot := []string{"a", "b"}
	currentState := []string{"a"}

	zkEvent := zk.findEvent(snapshot, currentState)
	if zkEvent.Event != DELETE_EVENT || zkEvent.ZNode != "b" {
		t.Fail()
	}

	snapshot, currentState = currentState, snapshot
	zkEvent = zk.findEvent(snapshot, currentState)
	if zkEvent.Event != CREATE_EVENT || zkEvent.ZNode != "b" {
		t.Fail()
	}
}
