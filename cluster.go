package client_announcer

import (
	"encoding/json"
	"errors"
	"fmt"
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
		if err := conn.Connect(nodeAdd, username, password, domain, time.Duration(cluster.Client.Duration)*time.Second); err != nil {
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
func (cluster *Cluster) SendToUsers(msgTemplate string, users []string) {
	for _, user := range users {
		msg := fmt.Sprintf(msgTemplate, user)
		if err := cluster.send(msg); err != nil {
			fmt.Sprintf("error: %s \n\ruser: %s \n\rmessage: %s", err.Error(), user, msg)
		}
	}
}

func (cluster *Cluster) send(msg string) error {
	for _, conn := range cluster.connections {
		if err := conn.Send(msg); err == nil {
			return nil
		}
	}

	return errors.New("can not send message")
}

func (cluster *Cluster) toJson() (result []byte) {
	result, _ = json.Marshal(cluster)
	return
}

func (cluster *Cluster) Close() {
	for _, conn := range cluster.connections {
		conn.Close()
	}
}
