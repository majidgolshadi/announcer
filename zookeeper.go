package client_announcer

import (
	"github.com/samuel/go-zookeeper/zk"
	"strings"
	"time"
)

type zookeeper struct {
	connection       *zk.Conn
	childrenSnapshot []string
}

type zkEvent struct {
	Event ZkEventType
	ZNode string
}

type ZkEventType int

const (
	DELETE_EVENT ZkEventType = 1
	CREATE_EVENT ZkEventType = 2
)

func zookeeperFactory(zookeeperAddress string) (zkobj *zookeeper, err error) {
	zkobj = &zookeeper{}

	zkServers := strings.Split(zookeeperAddress, ",")
	zkobj.connection, _, err = zk.Connect(zkServers, time.Minute)

	return
}

func (zkobj *zookeeper) save(zNode string, value []byte) (err error) {
	_, err = zkobj.connection.Create(zNode, value, int32(0), zk.WorldACL(zk.PermAll))
	return
}

func (zkobj *zookeeper) get(zNode string) (result []byte, err error) {
	result, _, err = zkobj.connection.Get(zNode)
	return
}

func (zkobj *zookeeper) childrenW(znode string) (chan *zkEvent, error) {
	c := make(chan *zkEvent)
	child, _, evCh, err := zkobj.connection.ChildrenW(znode)
	if err != nil {
		return c, err
	}

	zkobj.childrenSnapshot = child

	go func() {
	listen:
		for range evCh {
			child, _, evCh, err = zkobj.connection.ChildrenW(znode)
			c <- zkobj.findEvent(zkobj.childrenSnapshot, child)
			zkobj.childrenSnapshot = child
			if err != nil {
				println("children watcher error: ", err.Error())
				goto listen
			}
		}
	}()

	return c, nil
}

func (zkobj *zookeeper) findEvent(snapshot []string, currentState []string) *zkEvent {
	e := &zkEvent{}

	e.Event = CREATE_EVENT
	if len(snapshot) > len(currentState) {
		e.Event = DELETE_EVENT
	}

	e.ZNode = difference(currentState, snapshot)
	return e
}

func difference(slice1 []string, slice2 []string) string {
	for i := 0; i < 2; i++ {
		for _, s1 := range slice1 {
			found := false
			for _, s2 := range slice2 {
				if s1 == s2 {
					found = true
					break
				}
			}
			// String not found. We add it to return slice
			if !found {
				return s1
			}
		}

		// Swap the slices, only if it was the first loop
		if i == 0 {
			slice1, slice2 = slice2, slice1
		}
	}

	return ""
}

func (zkobj *zookeeper) close() {
	zkobj.connection.Close()
}
