package client_announcer

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type Cluster struct {
	Client      *Client
	Component   *Component
	connections []Sender
	Addresses   string
	RateLimit int
	throttle <-chan time.Time
}

type Client struct {
	Username     string
	Password     string
	Domain       string
	PingInterval int
}

type Component struct {
	Name         string
	Secret       string
	PingInterval int
}

func ClusterClientFactory(nodeAddresses string, username string, password string, domain string, pingIntervalDuration int, rateLimit int) (*Cluster, error) {
	cluster := &Cluster{
		Addresses: nodeAddresses,
		RateLimit: rateLimit,
		throttle: time.Tick(time.Second / time.Duration(rateLimit)),
		Client: &Client{
			Username:     username,
			Password:     password,
			Domain:       domain,
			PingInterval: pingIntervalDuration,
		},
	}

	ejabberdAddresses := strings.Split(nodeAddresses, ",")
	for _, nodeAdd := range ejabberdAddresses {
		conn := &ClientSender{}
		if err := conn.Connect(nodeAdd, username, password, domain, time.Duration(cluster.Client.PingInterval)*time.Second); err != nil {
			return nil, err
		}

		cluster.connections = append(cluster.connections, conn)
	}

	return cluster, nil
}

func ClusterComponentFactory(nodeAddresses string, name string, secret string, pingIntervalDuration int, rateLimit int) (*Cluster, error) {
	cluster := &Cluster{
		Addresses: nodeAddresses,
		RateLimit: rateLimit,
		throttle: time.Tick(time.Second / time.Duration(rateLimit)),
		Component: &Component{
			Name:         name,
			Secret:       secret,
			PingInterval: pingIntervalDuration,
		},
	}

	ejabberdAddresses := strings.Split(nodeAddresses, ",")
	for _, nodeAdd := range ejabberdAddresses {
		conn := &ComponentSender{}
		if err := conn.Connect(nodeAdd, name, secret, time.Duration(cluster.Component.PingInterval)*time.Second); err != nil {
			return nil, err
		}

		cluster.connections = append(cluster.connections, conn)
	}

	return cluster, nil
}

func (cluster *Cluster) SendToUsers(msgTemplate string, users []string) {
	for _, user := range users {
		<-cluster.throttle

		msg := fmt.Sprintf(msgTemplate, user)
		if err := cluster.send(msg); err != nil {
			fmt.Printf("error='%s' user='%s' message='%s' \n", err.Error(), user, msg)
		}
	}
}

func (cluster *Cluster) send(msg string) error {
	var err error
	for _, conn := range cluster.connections {
		if err = conn.Send(msg); err == nil {
			return nil
		}
	}

	return err
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
