package client_announcer

import (
	"strings"
	"errors"
	"time"
)

type Cluster struct {
	client *client
	component *component
	Connections []Connection

	sender string
}

type client struct {
	username string
	password string
	domain string
	duration int
}

type component struct {
	name string
	secret string
}

func ClusterClientFactory(username string, password string, domain string, connectionAlivePingDuration int, nodeAddresses string) (*Cluster, error) {
	cluster := &Cluster{
		client: &client{
			username: username,
			password: password,
			domain: domain,
			duration: connectionAlivePingDuration,
		},
	}

	nodesAddressArray := strings.Split(nodeAddresses, ",")

	for _,nodeAdd := range nodesAddressArray {
		conn := &ClientConnection{}
		if err := conn.Connect(nodeAdd, username, password, domain, time.Duration(cluster.client.duration)); err != nil {
			return nil, err
		}

		cluster.Connections = append(cluster.Connections, conn)
	}

	return cluster, nil
}

func ClusterComponentFactory(name string, secret string, nodeAddresses string) (*Cluster, error) {
	cluster := &Cluster{
		component: &component{
			name: name,
			secret: secret,
		},
	}

	nodesAddressArray := strings.Split(nodeAddresses, ",")

	for _,nodeAdd := range nodesAddressArray {
		conn := &ComponentConnection{}
		if err := conn.Connect(nodeAdd, name, secret); err != nil {
			return nil, err
		}

		cluster.Connections = append(cluster.Connections, conn)
	}

	return cluster, nil
}

// TODO: Add sending rate
func (cluster *Cluster)Send(msg string) error {
	for _, conn := range cluster.Connections {
		if err := conn.Send(msg); err == nil {
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
