package client_announcer

import (
	"errors"
	"strings"
	"time"
)

type Cluster struct {
	Client      *Client
	Component   *Component
	connections []Connection
	Addresses   string
}

type Client struct {
	Username string
	Password string
	Domain   string
	Duration int
}

type Component struct {
	Name   string
	Secret string
}

func ClusterClientFactory(username string, password string, domain string, connectionAlivePingDuration int, nodeAddresses string) (*Cluster, error) {
	cluster := &Cluster{
		Client: &Client{
			Username: username,
			Password: password,
			Domain:   domain,
			Duration: connectionAlivePingDuration,
		},
		Addresses: nodeAddresses,
	}

	nodesAddressArray := strings.Split(nodeAddresses, ",")

	for _, nodeAdd := range nodesAddressArray {
		conn := &ClientConnection{}
		if err := conn.Connect(nodeAdd, username, password, domain, time.Duration(cluster.Client.Duration)); err != nil {
			return nil, err
		}

		cluster.connections = append(cluster.connections, conn)
	}

	return cluster, nil
}

func ClusterComponentFactory(name string, secret string, nodeAddresses string) (*Cluster, error) {
	cluster := &Cluster{
		Component: &Component{
			Name:   name,
			Secret: secret,
		},
		Addresses: nodeAddresses,
	}

	nodesAddressArray := strings.Split(nodeAddresses, ",")

	for _, nodeAdd := range nodesAddressArray {
		conn := &ComponentConnection{}
		if err := conn.Connect(nodeAdd, name, secret); err != nil {
			return nil, err
		}

		cluster.connections = append(cluster.connections, conn)
	}

	return cluster, nil
}

// TODO: Add sending rate
func (cluster *Cluster) Send(msg string) error {
	for _, conn := range cluster.connections {
		if err := conn.Send(msg); err == nil {
			return nil
		}
	}

	return errors.New("can not send the message")
}

func (cluster *Cluster) Close() {
	for _, conn := range cluster.connections {
		conn.Close()
	}
}
