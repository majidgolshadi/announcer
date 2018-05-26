package output

import (
	"errors"
	"github.com/Shopify/sarama"
	log "github.com/sirupsen/logrus"
	"time"
)

type KafkaProducer struct {
	producer sarama.AsyncProducer
	config   *sarama.Config
	opt      *KafkaProducerOpt
}

type KafkaProducerOpt struct {
	Brokers        []string
	Topics         []string
	MaxRetry       int
	FlushFrequency time.Duration
}

func (opt *KafkaProducerOpt) init() error {
	if len(opt.Topics) < 1 {
		return errors.New("unknown topic")
	}

	if len(opt.Brokers) < 1 {
		opt.Brokers = []string{"127.0.0.1:9092"}
	}

	if opt.MaxRetry == 0 {
		opt.MaxRetry = 5
	}

	return nil
}

func NewKafkaProducer(option *KafkaProducerOpt) (*KafkaProducer, error) {
	if err := option.init(); err != nil {
		return nil, err
	}

	return &KafkaProducer{
		opt:    option,
		config: option.getSaramConfig(),
	}, nil
}

func (opt *KafkaProducerOpt) getSaramConfig() *sarama.Config {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForLocal
	config.Producer.Compression = sarama.CompressionNone
	config.Producer.Flush.Frequency = opt.FlushFrequency
	config.Producer.Retry.Max = opt.MaxRetry

	return config
}

func (kp *KafkaProducer) Listen(messages <-chan string) (err error) {
	kp.producer, err = sarama.NewAsyncProducer(kp.opt.Brokers, kp.config)
	if err != nil {
		return err
	}

	go kp.activeErrorEventListener()
	go kp.activeSuccessEventListener()
	go kp.action(messages)

	return nil
}

func (kp *KafkaProducer) action(messages <-chan string) {

	for message := range messages {
		for _, topic := range kp.opt.Topics {

			kp.producer.Input() <- &sarama.ProducerMessage{
				Topic: topic,
				Value: sarama.StringEncoder(message),
			}

		}
	}

}

func (kp *KafkaProducer) activeErrorEventListener() {
	for err := range kp.producer.Errors() {
		log.WithField("error", err).Error("kafka producer")
	}
}

func (kp *KafkaProducer) activeSuccessEventListener() {
	for msg := range kp.producer.Successes() {
		log.WithField("message", msg.Value).Debug("kafka producer message successfully sent")
	}
}

func (kp *KafkaProducer) Close() {
	log.Warn("kafka producer close")

	if kp.producer != nil {
		kp.producer.Close()
	}
}
