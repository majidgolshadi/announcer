package output

import (
	log "github.com/sirupsen/logrus"
	"time"
)

type Cluster struct {
	address []string
	conn    []Ejabberd
	retry   int
}

type Msg struct {
	Temp string
	User string
}

func NewClientCluster(address []string, sendRetry int, opt *EjabberdClientOpt) (c *Cluster, err error) {
	c = &Cluster{
		address: address,
		retry:   sendRetry,
	}

	for _, host := range address {
		client, err := NewEjabberdClient(&EjabberdClientOpt{
			Host:           host,
			Username:       opt.Username,
			Password:       opt.Password,
			PingInterval:   opt.PingInterval,
			ResourcePrefix: opt.ResourcePrefix,
			Domain:         opt.Domain,
		})

		if err != nil {
			return nil, err
		}

		err = client.Connect()
		if err != nil {
			return nil, err
		}

		c.conn = append(c.conn, client)
	}

	return c, nil
}

func NewComponentCluster(address []string, retry int, opt *EjabberdComponentOpt) (c *Cluster, err error) {
	c = &Cluster{
		address: address,
		retry:   retry,
	}

	for _, host := range address {
		com, err := NewEjabberdComponent(&EjabberdComponentOpt{
			Host:         host,
			Name:         opt.Name,
			Secret:       opt.Secret,
			PingInterval: opt.PingInterval,
			Domain:       opt.Domain,
		})

		if err != nil {
			return nil, err
		}

		err = com.Connect()
		if err != nil {
			return nil, err
		}

		c.conn = append(c.conn, com)
	}

	return c, nil
}

// Based on announcer usage it will be drop a message that it can't send, after retry on all connections and pause time
func (c *Cluster) ListenAndSend(rateLimit time.Duration, messages chan *Msg) {
	sleepTime := time.Second / rateLimit
	// For more that 1000000000 sleep time is zero
	if sleepTime == 0 {
		sleepTime = 1
	}

	ticker := time.NewTicker(sleepTime)

	for msg := range messages {
		<-ticker.C
		c.sendWithRetry(msg)
	}
}

func (c *Cluster) sendWithRetry(msg *Msg) {
	for i := 1; i < c.retry; i++ {
		for _, conn := range c.conn {
			if err := conn.Send(msg); err != nil {
				continue
			}

			return
		}

		log.WithField("message", msg).Warn("retry to send message after ", i, " second...")
		time.Sleep(time.Second * time.Duration(i))
	}
}

func (c *Cluster) Close() {
	log.Warn("close cluster connections...")
	for _, con := range c.conn {
		con.Close()
	}
}
