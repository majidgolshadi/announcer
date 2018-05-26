package input

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/majidgolshadi/client-announcer/logic"
	log "github.com/sirupsen/logrus"
	"github.com/wvanbergen/kafka/consumergroup"
	"math/rand"
	"time"
)

type KafkaConsumer struct {
	consumerGroup *consumergroup.ConsumerGroup
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
	template  []byte
}

func (kc *KafkaConsumer) Listen(inputChannel chan<- *logic.ChannelAct, inputUser chan<- *logic.UserAct) (err error) {
	if kc.consumerGroup, err = consumergroup.JoinConsumerGroup(kc.opt.GroupName, kc.opt.Topics, kc.opt.Zookeeper, kc.config); err != nil {
		return err
	}

	go kc.action(kc.consumerGroup.Messages(), inputChannel, inputUser)
	return nil
}

func (kc *KafkaConsumer) action(messageChannel <-chan *sarama.ConsumerMessage, inputChannel chan<- *logic.ChannelAct, inputUser chan<- *logic.UserAct) {
	for message := range messageChannel {
		req := &kafkaMsg{}
		if err := saramMessageUnmarshal(message, req); err != nil {
			log.WithField("error", err.Error()).
				Error("kafka input request")

			continue
		}

		if req.ChannelID != "" {
			inputChannel <- &logic.ChannelAct{
				MessageTemplate: string(req.template),
				ChannelID:       req.ChannelID,
			}
		} else {
			inputUser <- &logic.UserAct{
				MessageTemplate: string(req.template),
				Usernames:       req.Usernames,
			}
		}
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
	log.Warn("kafka input close")

	if kc.consumerGroup != nil {
		kc.consumerGroup.Close()
	}
}
