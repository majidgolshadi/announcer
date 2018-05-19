package input

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/Shopify/sarama"
	"github.com/majidgolshadi/client-announcer/logic"
	"github.com/majidgolshadi/client-announcer/output"
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
		opt.CommitOffsetInterval = 5
	}

	opt.CommitOffsetInterval = opt.CommitOffsetInterval * time.Second

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

func NewKafkaConsumer(option *KafkaConsumerOpt) (*KafkaConsumer, error) {
	if err := option.init(); err != nil {
		return nil, err
	}

	kc := &KafkaConsumer{
		opt:          option,
		commitTicker: time.NewTicker(option.CommitOffsetInterval),
	}
	kc.config = consumergroup.NewConfig()

	kc.config.ChannelBufferSize = option.ReadBufferSize
	kc.config.Offsets.Initial = sarama.OffsetNewest
	kc.config.Zookeeper.Chroot = option.ZNode
	kc.config.ChannelBufferSize = option.ReadBufferSize

	return kc, nil
}

type kafkaMsg struct {
	ChannelID string   `json:"channel_id"`
	Usernames []string `json:"usernames"`
	Message   string   `json:"message"`
}

func (kc *KafkaConsumer) Listen(inputChannel chan<- *logic.ChannelAct, outputChannel chan<- *output.Msg) (err error) {
	if kc.consumerGroup, err = consumergroup.JoinConsumerGroup(kc.opt.GroupName, kc.opt.Topics, kc.opt.Zookeeper, kc.config); err != nil {
		return err
	}

	go kc.tickOffsetCommitter()

	for message := range kc.consumerGroup.Messages() {
		req := &kafkaMsg{}
		if err := json.Unmarshal(message.Value, req); err != nil {
			log.Error("kafka request json unmarshal error: ", err.Error())
			continue
		}

		mstTemp, err := base64.StdEncoding.DecodeString(req.Message)
		if err != nil {
			log.Error("kafka request base64 convert error: ", err.Error())
			continue
		}

		if req.ChannelID != "" {
			inputChannel <- &logic.ChannelAct{
				MessageTemplate: string(mstTemp),
				ChannelID:       req.ChannelID,
			}
		} else {

			if len(req.Usernames) < 1 {
				log.Error("kafka request usernames len less than 1")
			}

			for _, user := range req.Usernames {
				outputChannel <- &output.Msg{
					Temp: string(mstTemp),
					User: user,
				}
			}

		}

		kc.lastMessage = message
	}

	return nil
}

func (kc *KafkaConsumer) tickOffsetCommitter() {
	for range kc.commitTicker.C {
		if kc.lastMessage != nil {
			kc.consumerGroup.CommitUpto(kc.lastMessage)
		}
	}
}

func (kc *KafkaConsumer) Close() {
	log.Warn("close kafka consumer")
	kc.commitTicker.Stop()

	if kc.consumerGroup != nil {
		kc.consumerGroup.CommitUpto(kc.lastMessage)
		kc.consumerGroup.Close()
	}
}
