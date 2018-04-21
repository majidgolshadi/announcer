package client_announcer

import (
	"encoding/json"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
)

type Cluster struct {
	Status      bool
	Client      ClientSender
	Component   ComponentSender
	connections map[string]Sender
	Addresses   []string
	RateLimit   int
}

func (cluster *Cluster) Connect() error {
	cluster.connections = make(map[string]Sender)
	for _, nodeAddr := range cluster.Addresses {
		sender := cluster.createSender()
		if err := sender.Connect(nodeAddr); err == nil {
			cluster.connections[nodeAddr] = sender
		}
	}

	cluster.Status = true
	return nil
}

func (cluster *Cluster) createSender() Sender {
	if cluster.Client.Password != "" {
		return &ClientSender{
			Username:     cluster.Client.Username,
			Password:     cluster.Client.Password,
			Domain:       cluster.Client.Domain,
			PingInterval: cluster.Client.PingInterval,
		}
	}

	return &ComponentSender{
		Name:         cluster.Component.Name,
		Secret:       cluster.Component.Secret,
		PingInterval: cluster.Component.PingInterval,
	}
}

func (cluster *Cluster) SendToUsers(msgTemplate string, users []string) error {
	sleepTime := time.Second / time.Duration(cluster.RateLimit)

	// For more that 1000000000 sleep time is zero
	if sleepTime == 0 {
		sleepTime = 1
	}

	ticker := time.NewTicker(sleepTime)

	for _, user := range users {
		<-ticker.C
		if cluster.Status {

			go func() {
				msg := fmt.Sprintf(msgTemplate, user)
				if err := cluster.send(msg); err != nil {
					log.WithFields(log.Fields{
						"error":   err.Error(),
						"user":    user,
						"message": msg}).Error("can't send message")
				}
			}()

		}
	}

	ticker.Stop()
	return nil
}

func (cluster *Cluster) send(msg string) error {
	var err error
	for key, conn := range cluster.connections {
		if err = conn.Send(msg); err == nil {
			return nil
		}

		delete(cluster.connections, key)
		log.Warn("connection to chat server", key, " lost")
	}

	go func() {
		if len(cluster.connections) < 1 {
			cluster.Status = false
			cluster.Connect()
		}
	}()

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
