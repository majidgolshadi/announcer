package client_announcer

import (
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/wvanbergen/kafka/consumergroup"
	"os"
	"os/signal"
	"time"
	"encoding/json"
)

var tickTimeMsg *sarama.ConsumerMessage

type KafkaConsumer struct {
	Zookeeper            []string
	ZNode                string
	GroupName            string
	Topics               []string
	Buffer               int
	CommitOffsetInterval time.Duration

	consumer          *consumergroup.ConsumerGroup
}

func (kc *KafkaConsumer) RunService(repository *ChatServerClusterRepository, onlineUserInquiry *onlineUserInquiry) error {
	c, err := kc.connect()
	if err != nil {
		return err
	}

	go func() {
		for msg := range c {
			recMsg := &kafkaMsg{}
			if err := json.Unmarshal(msg, recMsg); err != nil {
				println(err.Error())
			}

			if recMsg.ChannelID == "" && recMsg.Username == ""{
				println("bad request")
				continue
			}

			if recMsg.Cluster == "" {
				recMsg.Cluster = repository.defaultCluster
			}

			var users []string
			if recMsg.Username != "" {
				users = []string{recMsg.Username}
			} else {
				if users, err = onlineUserInquiry.GetOnlineUsers(recMsg.ChannelID); err != nil {
					println(err.Error())
					continue
				}
			}

			cluster, err := repository.Get(recMsg.Cluster)
			if err != nil {
				println(err.Error())
			}

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
