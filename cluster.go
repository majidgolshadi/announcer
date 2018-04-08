package client_announcer

import (
	"strings"
	"errors"
	"time"
	"fmt"
)

type Cluster struct {
	username string
	password string
	domain string
	duration int
	Connections []*NodeConnection

	sender string
}

func ClusterFactory(username string, password string, domain string, connectionAlivePingDuration int, nodeAddresses string) (*Cluster, error) {
	cluster := &Cluster{
		username: username,
		password: password,
		domain: domain,
		duration: connectionAlivePingDuration,
	}

	nodesAddressArray := strings.Split(nodeAddresses, ",")

	for _,nodeAdd := range nodesAddressArray {
		conn := &NodeConnection{}
		if err := conn.Connect(nodeAdd, username, password, domain, time.Duration(cluster.duration)); err != nil {
			return nil, err
		}

		cluster.Connections = append(cluster.Connections, conn)
		cluster.sender = fmt.Sprintf("%s@%s/%s", username, domain, "announcer")
	}

	return cluster, nil
}

func (cluster *Cluster)Send(msgTemplate string, to string) error {
	for _, conn := range cluster.Connections {
		if err := conn.Send(fmt.Sprintf(msgTemplate, to, cluster.sender)); err == nil {
			return nil
		}
	}

	return errors.New("can not send the message")
}

func (cluster *Cluster) Close() {
	for _, conn := range cluster.Connections {
		conn.Close()
	}
}
