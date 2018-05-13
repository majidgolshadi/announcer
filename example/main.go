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
	HttpPort     string `toml:"rest_api_port"`
	DebugPort    string `toml:"debug_port"`
	InputBuffer  int    `toml:"input_buffer"`
	OutputBuffer int    `toml:"output_buffer"`

	Log       Log
	Kafka     Kafka
	Ejabberd  Ejabberd
	Client    Client
	Component Component
	Mysql     Mysql
	Redis     Redis
}

type Kafka struct {
	Zookeeper            string `toml:"zookeeper"`
	Topics               string `toml:"topics"`
	GroupName            string `toml:"group_name"`
	Buffer               int    `toml:"buffer"`
	CommitOffsetInterval int    `toml:"commit_offset_interval"`
}

type Log struct {
	Format   string `toml:"format"`
	LogLevel string `toml:"log_level"`
	LogPoint string `toml:"log_point"`
}

type Ejabberd struct {
	ClusterNodes string `toml:"cluster_nodes"`
	RateLimit    int    `toml:"rate_limit"`
	SendRetry    int    `toml:"send_retry"`
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
	ClusterNodes  string `toml:"cluster_nodes"`
	Password      string `toml:"password"`
	DB            int    `toml:"db"`
	HashTable     string `toml:"hash_table"`
	CheckInterval int    `toml:"check_interval"`
}

type Mysql struct {
	Address       string `toml:"address"`
	Username      string `toml:"username"`
	Password      string `toml:"password"`
	DB            string `toml:"db"`
	CheckInterval int    `toml:"check_interval"`
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
	inputUser := make(chan *logic.UserAct, cnf.InputBuffer)
	out := make(chan *output.Msg, cnf.OutputBuffer)

	// Output part
	var cluster *output.Cluster
	if cnf.Component.Secret != "" {
		cluster, err = output.NewComponentCluster(strings.Split(cnf.Ejabberd.ClusterNodes, ","),
			cnf.Ejabberd.SendRetry,
			&output.EjabberdComponentOpt{
				Name:         cnf.Component.Name,
				Secret:       cnf.Component.Secret,
				PingInterval: time.Duration(cnf.Component.PingInterval),
				Domain:       cnf.Component.Domain,
			})
	} else {
		cluster, err = output.NewClientCluster(strings.Split(cnf.Ejabberd.ClusterNodes, ","),
			cnf.Ejabberd.SendRetry,
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

	go cluster.ListenAndSend(time.Duration(cnf.Ejabberd.RateLimit), out)

	// Logic part
	redis := logic.NewRedis(&logic.RedisOpt{
		Address:       cnf.Redis.ClusterNodes,
		Password:      cnf.Redis.Password,
		Database:      cnf.Redis.DB,
		CheckInterval: time.Duration(cnf.Redis.CheckInterval),
	})

	mysql, err := logic.NewMysql(&logic.MysqlOpt{
		Address:       cnf.Mysql.Address,
		Database:      cnf.Mysql.DB,
		Username:      cnf.Mysql.Username,
		Password:      cnf.Mysql.Password,
		CheckInterval: time.Duration(cnf.Mysql.CheckInterval),
	})
	if err != nil {
		log.WithField("error", err.Error()).Fatal("mysql connection failed")
	}

	chActor := &logic.ChannelActor{
		Mysql:          mysql,
		Redis:          redis,
		RedisHashTable: cnf.Redis.HashTable,
	}

	go chActor.Listen(inputChannel, out)
	defer chActor.Close()

	usActor := &logic.UserActor{}
	go usActor.Listen(inputUser, out)

	// Input part
	// Kafka consumer
	var kafkaConsumer *input.KafkaConsumer
	if cnf.Kafka.Zookeeper != "" {
		zookeeper, zNode := kazoo.ParseConnectionString(cnf.Kafka.Zookeeper)
		kafkaConsumer, err = input.NewKafkaConsumer(&input.KafkaConsumerOpt{
			Zookeeper:            zookeeper,
			ZNode:                zNode,
			GroupName:            cnf.Kafka.GroupName,
			CommitOffsetInterval: time.Duration(cnf.Kafka.CommitOffsetInterval),
			Topics:               strings.Split(cnf.Kafka.Topics, ","),
			ReadBufferSize:       cnf.Kafka.Buffer,
		})

		defer kafkaConsumer.Close()

		if err != nil {
			log.WithField("error", err.Error()).Fatal("init kafka consumer failed")
		}

		go func() {
			if err := kafkaConsumer.Listen(inputChannel, inputUser); err != nil {
				log.WithField("error", err.Error()).Fatal("kafka consumer listening error")
			}
		}()
	}

	// Rest api
	input.RunHttpServer(cnf.HttpPort, inputChannel, inputUser)
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
	case "text":
		log.SetFormatter(&log.TextFormatter{})
	default:
		break
	}

	if logConfig.LogPoint != "" {
		f, err := os.Create(logConfig.LogPoint)
		if err != nil {
			log.Fatal("create log file error: ", err.Error())
		}

		log.SetOutput(f)
	}
}
