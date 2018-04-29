package client_announcer

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/wvanbergen/kafka/consumergroup"
	"os"
	"os/signal"
	"time"
)

var tickTimeMsg *sarama.ConsumerMessage

type KafkaConsumer struct {
	Zookeeper            []string
	ZNode                string
	GroupName            string
	Topics               []string
	Buffer               int
	CommitOffsetInterval time.Duration

	consumer *consumergroup.ConsumerGroup
}

type kafkaMsg struct {
	ChannelID string `json:"channel_id"`
	Username  string `json:"username"`
	Message   string `json:"message"`
	Cluster   string `json:"cluster"`
}

func (kc *KafkaConsumer) getUsers(onlineUserInquiry *onlineUserInquiry, req *kafkaMsg) (users []string, err error) {
	if req.ChannelID == "" && req.Username == "" {
		return nil, errors.New("channel_id or username is empty")
	}

	users = []string{req.Username}
	if req.ChannelID != "" {
		users, err = onlineUserInquiry.GetOnlineUsers(req.ChannelID)
	}

	return users, nil
}

func (kc *KafkaConsumer) RunService(repo *ChatServerClusterRepository, onlineUserInquiry *onlineUserInquiry) error {
	c, err := kc.connect()
	if err != nil {
		return err
	}

	go func() {
		var users []string
		var cluster *Cluster
		recMsg := &kafkaMsg{}

		for req := range c {
			if err := json.Unmarshal(req, recMsg); err != nil {
				println(err.Error())
				continue
			}

			if users, err = kc.getUsers(onlineUserInquiry, recMsg); err != nil {
				println(err.Error())
				continue
			}

			cluster, _ = repo.Get(recMsg.Cluster)
			cluster.SendToUsers(recMsg.Message, users)
		}
	}()

	return nil
}

func (kc *KafkaConsumer) connect() (msgChan chan []byte, err error) {
	if kc.consumer, err = consumergroup.JoinConsumerGroup(kc.GroupName, kc.Topics, kc.Zookeeper, kc.getConfig()); err != nil {
		return nil, err
	}

	kc.setupInterruptListener()
	go kc.tickOffsetCommitter()

	go func() {
		for message := range kc.consumer.Messages() {
			tickTimeMsg = message
			msgChan <- message.Value
		}
	}()

	return msgChan, nil
}

func (kc *KafkaConsumer) getConfig() *consumergroup.Config {
	config := consumergroup.NewConfig()
	config.ChannelBufferSize = 100
	config.Offsets.Initial = sarama.OffsetOldest
	config.Zookeeper.Chroot = kc.ZNode

	if kc.Buffer != 0 {
		config.ChannelBufferSize = kc.Buffer
	}

	return config
}

func (kc *KafkaConsumer) setupInterruptListener() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		<-c
		println("Os interrupt signal received")

		if err := kc.consumer.Close(); err != nil {
			fmt.Printf("Error closing the consumer %s \n", err.Error())
		}
	}()
}

func (kc *KafkaConsumer) tickOffsetCommitter() {
	commitTicker := time.NewTicker(kc.CommitOffsetInterval * time.Millisecond)
	defer commitTicker.Stop()

	for range commitTicker.C {
		if tickTimeMsg != nil {
			kc.consumer.CommitUpto(tickTimeMsg)
		}
	}
}

func (kc *KafkaConsumer) Close() {
	kc.consumer.Close()
}
