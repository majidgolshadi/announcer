package client_announcer

import (
	"encoding/json"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
)

type Cluster struct {
	Client      *Client
	Component   *Component
	connections []Sender
	Addresses   []string
	RateLimit   int
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

func ClusterClientFactory(nodeAddresses []string, username string, password string, domain string, pingIntervalDuration int, rateLimit int) (*Cluster, error) {
	cluster := &Cluster{
		Addresses: nodeAddresses,
		RateLimit: rateLimit,
		Client: &Client{
			Username:     username,
			Password:     password,
			Domain:       domain,
			PingInterval: pingIntervalDuration,
		},
	}

	for _, nodeAdd := range cluster.Addresses {
		conn := &ClientSender{}
		if err := conn.Connect(nodeAdd, username, password, domain, time.Duration(cluster.Client.PingInterval)*time.Second); err != nil {
			return nil, err
		}

		cluster.connections = append(cluster.connections, conn)
	}

	return cluster, nil
}

func ClusterComponentFactory(nodeAddresses []string, name string, secret string, pingIntervalDuration int, rateLimit int) (*Cluster, error) {
	cluster := &Cluster{
		Addresses: nodeAddresses,
		RateLimit: rateLimit,
		Component: &Component{
			Name:         name,
			Secret:       secret,
			PingInterval: pingIntervalDuration,
		},
	}

	for _, nodeAdd := range cluster.Addresses {
		conn := &ComponentSender{}
		if err := conn.Connect(nodeAdd, name, secret, time.Duration(cluster.Component.PingInterval)*time.Second); err != nil {
			return nil, err
		}

		cluster.connections = append(cluster.connections, conn)
	}

	return cluster, nil
}

func (cluster *Cluster) SendToUsers(msgTemplate string, users []string) {
	sleepTime := time.Second / time.Duration(cluster.RateLimit)

	if sleepTime == 0 {
		sleepTime = 1
	}

	ticker := time.NewTicker(sleepTime)

	for _, user := range users {
		<-ticker.C

		msg := fmt.Sprintf(msgTemplate, user)
		if err := cluster.send(msg); err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
				"user": user,
				"message": msg}).Error("can't send message")
		}
	}

	ticker.Stop()
}

func (cluster *Cluster) send(msg string) error {
	var err error
	for index, conn := range cluster.connections {
		if err = conn.Send(msg); err == nil {
			return nil
		}

		log.Error("send message to server ", cluster.Addresses[index], " failed")
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
