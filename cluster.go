package client_announcer

import (
	"encoding/json"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	"sync"
)

type Cluster struct {
	Client      ClientSender
	Component   ComponentSender
	connections map[string]Sender
	domain      string
	Addresses   []string
	RateLimit   int
	SendRetry   int
	ticker      *time.Ticker
	connectionMutex *sync.Mutex
}

func (cluster *Cluster) Run() error {
	cluster.connectionMutex = &sync.Mutex{}
	cluster.initRateLimitTicker()
	return cluster.connect()
}

func (cluster *Cluster) initRateLimitTicker() {
	sleepTime := time.Second / time.Duration(cluster.RateLimit)
	// For more that 1000000000 sleep time is zero
	if sleepTime == 0 {
		sleepTime = 1
	}

	cluster.ticker = time.NewTicker(sleepTime)
}

func (cluster *Cluster) connect() error {
	cluster.connections = make(map[string]Sender)
	for _, nodeAddr := range cluster.Addresses {
		sender := cluster.createSender()
		if err := sender.Connect(nodeAddr); err == nil {
			cluster.connections[nodeAddr] = sender
		}
	}

	return nil
}

func (cluster *Cluster) createSender() Sender {
	if cluster.Client.Password != "" {
		cluster.domain = cluster.Client.Domain
		return &ClientSender{
			Username:     cluster.Client.Username,
			Password:     cluster.Client.Password,
			Domain:       cluster.Client.Domain,
			Resource:     cluster.Client.Resource,
			PingInterval: cluster.Client.PingInterval,
		}
	}

	cluster.domain = cluster.Component.Domain
	return &ComponentSender{
		Name:         cluster.Component.Name,
		Secret:       cluster.Component.Secret,
		PingInterval: cluster.Component.PingInterval,
		Domain:       cluster.Component.Domain,
	}
}

func (cluster *Cluster) SendToUsers(msgTemplate string, users []string) error {
	for _, user := range users {
		<-cluster.ticker.C
		msg := fmt.Sprintf(msgTemplate, fmt.Sprintf("%s@%s", user, cluster.domain))
		go func() {
			for i := 1; i < cluster.SendRetry; i++ {
				if err := cluster.send(msg); err == nil {
					log.Info(msg, " sent")
					return
				}
				timer := time.NewTimer(time.Duration(i) * time.Second)
				log.WithField("message", msg).Error("retry to send message")
				<-timer.C
			}
		}()
	}

	return nil
}

func (cluster *Cluster) send(msg string) error {
	var err error
	for key, conn := range cluster.connections {
		if err = conn.Send(msg); err == nil {
			return nil
		}

		log.Warn("connection to chat server", key, " lost")
		conn.Close()
		delete(cluster.connections, key)
	}

	cluster.connectionMutex.Lock()
	if len(cluster.connections) < 1 {
		cluster.connect()
	}
	cluster.connectionMutex.Unlock()

	return err
}

func (cluster *Cluster) toJson() (result []byte) {
	result, _ = json.Marshal(cluster)
	return
}

func (cluster *Cluster) Close() {
	log.Info("close cluster connection ", cluster.Addresses)
	cluster.ticker.Stop()
	for _, conn := range cluster.connections {
		conn.Close()
	}
}
