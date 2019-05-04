package input

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/majidgolshadi/client-announcer/logic"
	"github.com/majidgolshadi/client-announcer/output"
	log "github.com/sirupsen/logrus"
	"github.com/wvanbergen/kafka/consumergroup"
	"math/rand"
	"time"
	"strings"
)

type KafkaConsumer struct {
	consumerGroups []*consumergroup.ConsumerGroup
	config        *consumergroup.Config
	opt           *KafkaConsumerOpt

	lastMessage  *sarama.ConsumerMessage
	commitTicker *time.Ticker
}

type KafkaConsumerOpt struct {
	Zookeeper            []string
	ZNode                string
	GroupName            string
	Topics               []string
	ReadBufferSize       int
	CommitOffsetInterval time.Duration
}

func (opt *KafkaConsumerOpt) init() error {
	if len(opt.Topics) < 1 {
		return errors.New("unknown topic")
	}

	if len(opt.Zookeeper) < 1 {
		opt.Zookeeper = []string{"127.0.0.1:2181"}
	}

	if opt.ZNode == "/" {
		opt.ZNode = ""
	}

	if opt.GroupName == "" {
		opt.GroupName = generateGroupName(5)
	}

	if opt.ReadBufferSize == 0 {
		opt.ReadBufferSize = 100
	}

	if opt.CommitOffsetInterval == 0 {
		opt.CommitOffsetInterval = time.Second
	}

	return nil
}

func generateGroupName(n int) string {
	letterRunes := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return "AnnouncerConsumer_" + string(b)
}

func (opt *KafkaConsumerOpt) getSaramaConfig() *consumergroup.Config {
	config := consumergroup.NewConfig()
	config.ChannelBufferSize = opt.ReadBufferSize
	config.Offsets.Initial = sarama.OffsetNewest
	config.Offsets.CommitInterval = opt.CommitOffsetInterval
	config.Zookeeper.Chroot = opt.ZNode
	config.ChannelBufferSize = opt.ReadBufferSize

	return config
}

func NewKafkaConsumer(option *KafkaConsumerOpt) (*KafkaConsumer, error) {
	if err := option.init(); err != nil {
		return nil, err
	}

	kc := &KafkaConsumer{
		opt:    option,
		config: option.getSaramaConfig(),
	}

	return kc, nil
}

type kafkaMsg struct {
	ChannelID string   `json:"channel_id"`
	Usernames []string `json:"usernames"`
	Message   string   `json:"message"`
	Persist   bool     `json:"persist"`
	template  []byte
}

func (kc *KafkaConsumer) Listen(inputChannel chan<- *logic.ChannelAct, inputChat chan<- *output.Message) (err error) {
	for index, topic := range kc.opt.Topics {
		groupName := kc.opt.GroupName
		if index > 0 {
			groupName = fmt.Sprintf("%s_%d", groupName, index)
		}

		consumer, err := consumergroup.JoinConsumerGroup(groupName, []string{topic}, kc.opt.Zookeeper, kc.config)
		if err != nil {
			return err
		}

		kc.consumerGroups = append(kc.consumerGroups, consumer)
		log.Info("kafka consumer connection established with group name ", groupName, " on topic ", topic)

		go kc.action(index, consumer.Messages(), inputChannel, inputChat)
	}

	return nil
}

func (kc *KafkaConsumer) action(index int, messageChannel <-chan *sarama.ConsumerMessage, inputChannel chan<- *logic.ChannelAct, inputChat chan<- *output.Message) {
	for message := range messageChannel {
		req := &kafkaMsg{
			Persist: true,
		}
		if err := saramMessageUnmarshal(message, req); err != nil {
			log.WithField("error", err.Error()).
				Error("kafka consumer request")

			continue
		}

		if req.ChannelID != "" {
			inputChannel <- &logic.ChannelAct{
				MessageTemplate: string(req.template),
				ChannelID:       req.ChannelID,
			}
		} else {
			go func(kafkaMsg) {
				for index, username := range req.Usernames {

					templateMsg := string(req.template)

					if strings.Contains(string(req.template), "[inc_id]") {
						templateMsg = strings.Replace(templateMsg, "[inc_id]", string(index), -1)
					}

					inputChat <- &output.Message{
						Template: templateMsg,
						Username: username,
						Persist:  req.Persist,
					}
				}
			}(*req)
		}

		kc.consumerGroups[index].CommitUpto(message)
	}
}

func saramMessageUnmarshal(message *sarama.ConsumerMessage, msg *kafkaMsg) (err error) {
	if err = json.Unmarshal(message.Value, msg); err != nil {
		return errors.New(fmt.Sprintf("json unmarshal error %s", err.Error()))
	}

	if msg.Message == "" {
		return errors.New("bad request")
	}

	msg.template, err = base64.StdEncoding.DecodeString(msg.Message)
	if err != nil {
		return errors.New(fmt.Sprintf("base64 convert error %s", err.Error()))
	}

	if msg.ChannelID == "" && len(msg.Usernames) < 1 ||
		msg.ChannelID != "" && len(msg.Usernames) > 1 {
		return errors.New("bad request")
	}

	return nil
}

func (kc *KafkaConsumer) Close() {
	log.Warn("kafka consumer close")

	if kc.consumerGroups != nil {
		for _,consumer := range kc.consumerGroups {
			consumer.Close()
		}
	}
}
