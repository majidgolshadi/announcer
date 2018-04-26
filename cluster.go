package client_announcer

import (
	"encoding/json"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	"sync"
)

type Cluster struct {
	Client          ClientSender
	Component       ComponentSender
	connections     map[string]Sender
	domain          string
	Addresses       []string
	RateLimit       int
	SendRetry       int
	ticker          *time.Ticker
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
	log.Info("ticker init with sleep time ", sleepTime)
}

func (cluster *Cluster) connect() (err error) {
	cluster.connections = make(map[string]Sender)
	for _, nodeAddr := range cluster.Addresses {
		sender := cluster.createSenderInstance()
		if err = sender.Connect(nodeAddr); err == nil {
			cluster.connections[nodeAddr] = sender
		} else {
			log.WithField("error", err.Error()).Error("connecting to ", nodeAddr, " error")
		}
	}

	return
}

func (cluster *Cluster) createSenderInstance() Sender {
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

func (cluster *Cluster) SendToUsers(msgTemplate string, users []string) {
	log.Info("users to send ", len(users))
	for _, user := range users {
		<-cluster.ticker.C
		cluster.SendToUser(msgTemplate, user)
	}
}

func (cluster *Cluster) SendToUser(msgTemplate string, user string) {
	msg := fmt.Sprintf(msgTemplate, fmt.Sprintf("%s@%s", user, cluster.domain))
	go func() {
		cluster.sendWithRetry(msg)
	}()
}

func (cluster *Cluster) sendWithRetry(msg string) {
	for i := 1; i < cluster.SendRetry; i++ {
		if err := cluster.sendMsg(msg); err == nil {
			log.WithField("message", msg).Info("message sent")
			return
		}
		log.WithField("message", msg).Warn("retry to send message after ", i, " second...")
		time.Sleep(time.Second * time.Duration(i))
	}
}

func (cluster *Cluster) sendMsg(msg string) error {
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
