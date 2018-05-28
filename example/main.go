package main

import (
	"net/http"
	_ "net/http/pprof"
	"os"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/majidgolshadi/client-announcer/input"
	"github.com/majidgolshadi/client-announcer/logic"
	"github.com/majidgolshadi/client-announcer/output"
	log "github.com/sirupsen/logrus"
	"github.com/wvanbergen/kazoo-go"
)

type config struct {
	HttpPort             string `toml:"rest_api_port"`
	DebugPort            string `toml:"debug_port"`
	LogicProcessNum      int    `toml:"logic_process_number"`
	InputBuffer          int    `toml:"input_buffer"`
	OutputBuffer         int    `toml:"output_buffer"`
	BufferReportDuration int    `toml:"buffer_report_duration"`

	Log           Log
	Mysql         Mysql
	Redis         Redis
	Ejabberd      Ejabberd
	Client        Client
	Component     Component
	KafkaConsumer KafkaConsumer `toml:"kafka-consumer"`
	KafkaProducer KafkaProducer `toml:"kafka-producer"`
}

type KafkaConsumer struct {
	Zookeeper            string `toml:"zookeeper"`
	Topics               string `toml:"topics"`
	GroupName            string `toml:"group_name"`
	Buffer               int    `toml:"buffer"`
	CommitOffsetInterval int    `toml:"commit_offset_interval"`
}

type KafkaProducer struct {
	Brokers        string `toml:"brokers"`
	Topics         string `toml:"topics"`
	FlushFrequency int    `toml:"flush_frequency"`
	MaxRetry       int    `toml:"max_retry"`
}

type Log struct {
	Format   string `toml:"format"`
	LogLevel string `toml:"log_level"`
	LogPoint string `toml:"log_point"`
}

type Ejabberd struct {
	ClusterNodes    string `toml:"cluster_nodes"`
	RateLimit       int    `toml:"rate_limit"`
	SendRetry       int    `toml:"send_retry"`
	EachNodeConnNum int    `toml:"each_node_conn_num"`
}

type Client struct {
	Username     string `toml:"username"`
	Password     string `toml:"password"`
	Domain       string `toml:"domain"`
	Resource     string `toml:"resource"`
	PingInterval int    `toml:"ping_interval"`
}

type Component struct {
	Name         string `toml:"name"`
	Secret       string `toml:"secret"`
	Domain       string `toml:"domain"`
	PingInterval int    `toml:"ping_interval"`
}

type Redis struct {
	ClusterNodes string `toml:"cluster_nodes"`
	Password     string `toml:"password"`
	DB           int    `toml:"db"`
	SetPrefix    string `toml:"set_prefix"`
	ReadTimeout  int    `toml:"read_timeout"`
	MaxRetries   int    `toml:"max_retries"`
	MGetBuffer   int    `toml:"mget_buffer"`
}

type Mysql struct {
	Address            string `toml:"address"`
	Username           string `toml:"username"`
	Password           string `toml:"password"`
	DB                 string `toml:"db"`
	CheckInterval      int    `toml:"check_interval"`
	PaginationLength   int    `toml:"pagination_length"`
	MaxIdealConnection int    `toml:"max_ideal_conn"`
	MaxOpenConnection  int    `toml:"max_open_conn"`
}

func main() {
	var cnf config
	var err error

	if _, err := toml.DecodeFile("config.toml", &cnf); err != nil {
		log.Fatal("read configuration file error ", err.Error())
	}

	initLogService(cnf.Log)

	go func() {
		log.Info("debugging server listening on port ", cnf.DebugPort)
		log.Println(http.ListenAndServe(cnf.DebugPort, nil))
	}()

	inputChannel := make(chan *logic.ChannelAct, cnf.InputBuffer)
	defer close(inputChannel)

	inChat := make(chan *output.Message, cnf.OutputBuffer)
	defer close(inChat)

	inKafka := make(chan string, cnf.OutputBuffer)
	defer close(inKafka)

	///////////////////////////////////////////////////////////
	// Report
	///////////////////////////////////////////////////////////
	go func() {
		for {
			time.Sleep(time.Second * time.Duration(cnf.BufferReportDuration))
			log.WithFields(log.Fields{
				"input":    len(inputChannel),
				"outChat":  len(inChat),
				"outKafka": len(inKafka),
			}).Info("channel fill length")
		}
	}()

	///////////////////////////////////////////////////////////
	///////////////////////////////////////////////////////////
	// Output part
	///////////////////////////////////////////////////////////
	///////////////////////////////////////////////////////////

	///////////////////////////////////////////////////////////
	// Ejabberd Component
	///////////////////////////////////////////////////////////
	var cluster *output.Cluster
	if cnf.Component.Secret != "" {
		cluster, err = output.NewComponentCluster(
			strings.Split(cnf.Ejabberd.ClusterNodes, ","),
			cnf.Ejabberd.SendRetry,
			cnf.Ejabberd.EachNodeConnNum,
			&output.EjabberdComponentOpt{
				Name:         cnf.Component.Name,
				Secret:       cnf.Component.Secret,
				PingInterval: time.Duration(cnf.Component.PingInterval),
				Domain:       cnf.Component.Domain,
			})
	} else {

		///////////////////////////////////////////////////////////
		// Ejabberd Client
		///////////////////////////////////////////////////////////
		cluster, err = output.NewClientCluster(
			strings.Split(cnf.Ejabberd.ClusterNodes, ","),
			cnf.Ejabberd.SendRetry,
			cnf.Ejabberd.EachNodeConnNum,
			&output.EjabberdClientOpt{
				Username:       cnf.Client.Username,
				Password:       cnf.Client.Password,
				PingInterval:   time.Duration(cnf.Component.PingInterval),
				Domain:         cnf.Client.Domain,
				ResourcePrefix: cnf.Client.Domain,
			})
	}

	if err != nil {
		log.WithField("error", err.Error()).Fatal("cluster connecting error")
	}

	go cluster.ListenAndSend(cnf.Ejabberd.RateLimit, inChat, inKafka)
	defer cluster.Close()

	///////////////////////////////////////////////////////////
	// Kafka Producer
	///////////////////////////////////////////////////////////
	kafkaOpt := &output.KafkaProducerOpt{
		Brokers:        strings.Split(cnf.KafkaProducer.Brokers, ","),
		Topics:         strings.Split(cnf.KafkaProducer.Topics, ","),
		FlushFrequency: time.Duration(cnf.KafkaProducer.FlushFrequency),
		MaxRetry:       cnf.KafkaProducer.MaxRetry,
	}

	kafkaProducer, err := output.NewKafkaProducer(kafkaOpt)
	if err != nil {
		log.WithField("error", err.Error()).Fatal("kafka producer configuration error")
	}

	if err := kafkaProducer.Listen(inKafka); err != nil {
		log.WithField("error", err.Error()).Fatal("kafka producer connection error")
	}

	defer kafkaProducer.Close()

	///////////////////////////////////////////////////////////
	///////////////////////////////////////////////////////////
	// Logic part
	///////////////////////////////////////////////////////////
	///////////////////////////////////////////////////////////

	// Redis configuration
	redis, err := logic.NewRedisUserDataStore(&logic.RedisOpt{
		Address:     cnf.Redis.ClusterNodes,
		Password:    cnf.Redis.Password,
		Database:    cnf.Redis.DB,
		SetPrefix:   cnf.Redis.SetPrefix,
		MaxRetries:  cnf.Redis.MaxRetries,
		ReadTimeout: time.Duration(cnf.Redis.ReadTimeout),
	})
	if err != nil {
		log.WithField("error", err.Error()).Fatal("redis connection failed")
	}

	// Mysql configuration
	mysql, err := logic.NewMysqlChannelDataStore(&logic.MysqlOpt{
		Address:       cnf.Mysql.Address,
		Database:      cnf.Mysql.DB,
		Username:      cnf.Mysql.Username,
		Password:      cnf.Mysql.Password,
		CheckInterval: time.Duration(cnf.Mysql.CheckInterval),
		PageLength:    cnf.Mysql.PaginationLength,
		MaxIdealConn:  cnf.Mysql.MaxIdealConnection,
		MaxOpenConn:   cnf.Mysql.MaxOpenConnection,
	})
	if err != nil {
		log.WithField("error", err.Error()).Fatal("mysql connection failed")
	}

	if cnf.LogicProcessNum < 1 {
		log.Fatal("logic process number is less than 1")
	}

	///////////////////////////////////////////////////////////
	// Channel actor
	///////////////////////////////////////////////////////////
	var logicProcesses []*logic.ChannelActor

	for i := 0; i < cnf.LogicProcessNum; i++ {
		channelActor := &logic.ChannelActor{
			ChannelDataStore: mysql,
			UserActivity:     redis,
		}

		logicProcesses = append(logicProcesses, channelActor)

		go channelActor.Listen(cnf.Redis.MGetBuffer, inputChannel, inChat)
	}

	defer func() {
		for index := range logicProcesses {
			logicProcesses[index].Close()
		}
	}()

	///////////////////////////////////////////////////////////
	///////////////////////////////////////////////////////////
	// Input part
	///////////////////////////////////////////////////////////
	///////////////////////////////////////////////////////////

	///////////////////////////////////////////////////////////
	// Kafka
	///////////////////////////////////////////////////////////
	var kafkaConsumer *input.KafkaConsumer
	if cnf.KafkaConsumer.Zookeeper != "" {
		zookeeper, zNode := kazoo.ParseConnectionString(cnf.KafkaConsumer.Zookeeper)
		kafkaConsumer, err = input.NewKafkaConsumer(&input.KafkaConsumerOpt{
			Zookeeper:            zookeeper,
			ZNode:                zNode,
			GroupName:            cnf.KafkaConsumer.GroupName,
			Topics:               strings.Split(cnf.KafkaConsumer.Topics, ","),
			ReadBufferSize:       cnf.KafkaConsumer.Buffer,
			CommitOffsetInterval: time.Duration(cnf.KafkaConsumer.CommitOffsetInterval) * time.Second,
		})

		defer kafkaConsumer.Close()

		if err != nil {
			log.WithField("error", err.Error()).Fatal("init kafka consumer failed")
		}

		if err := kafkaConsumer.Listen(inputChannel, inChat); err != nil {
			log.WithField("error", err.Error()).Fatal("kafka consumer listening error")
		}
	}

	///////////////////////////////////////////////////////////
	// HTTP Rest API
	///////////////////////////////////////////////////////////
	input.RunHttpServer(cnf.HttpPort, inputChannel, inChat)
}

// TODO: Add tag for any application log
func initLogService(logConfig Log) {
	switch logConfig.LogLevel {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	default:
		log.SetLevel(log.WarnLevel)
	}

	switch logConfig.Format {
	case "json":
		log.SetFormatter(&log.JSONFormatter{})
	default:
		log.SetFormatter(&log.TextFormatter{})
	}

	if logConfig.LogPoint != "" {
		f, err := os.Create(logConfig.LogPoint)
		if err != nil {
			log.Fatal("create log file error: ", err.Error())
		}

		log.SetOutput(f)
	}
}
