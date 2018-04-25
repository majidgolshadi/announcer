package main

import (
	"net/http"
	_ "net/http/pprof"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/majidgolshadi/client-announcer"
	log "github.com/sirupsen/logrus"
)

type config struct {
	HttpPort  string `toml:"rest_api_port"`
	DebugPort string `toml:"debug_port"`

	Log       Log
	Ejabberd  Ejabberd
	Client    Client
	Component Component
	Zookeeper Zookeeper
	Mysql     Mysql
	Redis     Redis
}

type Log struct {
	Format   string `toml:"format"`
	LogLevel string `toml:"log_level"`
	LogPoint string `toml:"log_point"`
}

type Ejabberd struct {
	ClusterNodes   string `toml:"cluster_nodes"`
	DefaultCluster string `toml:"default_cluster"`
	RateLimit      int    `toml:"rate_limit"`
	SendRetry      int    `toml:"send_retry"`
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

type Zookeeper struct {
	ClusterNodes string `toml:"cluster_nodes"`
	NameSpace    string `toml:"namespace"`
}

type Redis struct {
	ClusterNodes  string `toml:"cluster_nodes"`
	Password      string `toml:"password"`
	DB            int    `toml:"db"`
	HashTable     string `toml:"hash_table"`
	CheckInterval int    `toml:"check_interval"`
}

type Mysql struct {
	Address  string `toml:"address"`
	Username string `toml:"username"`
	Password string `toml:"password"`
	DB       string `toml:"db"`
}

func main() {
	var (
		cnf          config
		cluster      *client_announcer.Cluster
		chatConnRepo *client_announcer.ChatServerClusterRepository
		err          error
	)

	if _, err := toml.DecodeFile("config.toml", &cnf); err != nil {
		log.Fatal("read configuration file error ", err.Error())
	}

	initLogService(cnf.Log)

	go func() {
		log.Info("debugging server listening on port ", cnf.DebugPort)
		log.Println(http.ListenAndServe(cnf.DebugPort, nil))
	}()

	if chatConnRepo, err = client_announcer.ChatServerClusterRepositoryFactory(cnf.Ejabberd.DefaultCluster); err != nil {
		log.Fatal("create repo error ", err.Error())
	}

	if cnf.Zookeeper.ClusterNodes != "" {
		if err = chatConnRepo.SetZookeeperAsDataStore(cnf.Zookeeper.ClusterNodes, cnf.Zookeeper.NameSpace); err != nil {
			log.Fatal("set zookeeper as data-store error ", err.Error())
			return
		}
	}

	defer chatConnRepo.Close()

	cluster = &client_announcer.Cluster{
		Addresses: strings.Split(cnf.Ejabberd.ClusterNodes, ","),
		RateLimit: cnf.Ejabberd.RateLimit,
		SendRetry: cnf.Ejabberd.SendRetry,
	}

	if cnf.Component.Secret != "" {
		cluster.Component.Name = cnf.Component.Name
		cluster.Component.Secret = cnf.Component.Secret
		cluster.Component.PingInterval = cnf.Component.PingInterval
		cluster.Component.Domain = cnf.Component.Domain

	} else if cnf.Client.Password != "" {
		cluster.Client.Username = cnf.Client.Username
		cluster.Client.Password = cnf.Client.Password
		cluster.Client.Domain = cnf.Client.Domain
		cluster.Client.Resource = cnf.Client.Resource
		cluster.Client.PingInterval = cnf.Client.PingInterval
	}

	if err = cluster.Run(); err != nil {
		log.Fatal("erjaberd create connection error ", err)
		return
	}

	err = chatConnRepo.Save(cnf.Ejabberd.DefaultCluster, cluster)
	if err != nil {
		log.Fatal("store in repository error ", err.Error())
		return
	}

	onlineUserInquiry, _ := client_announcer.OnlineUserInquiryFactory(
		cnf.Mysql.Address, cnf.Mysql.Username, cnf.Mysql.Password, cnf.Mysql.DB,
		cnf.Redis.ClusterNodes, cnf.Redis.Password, cnf.Redis.DB, cnf.Redis.HashTable, cnf.Redis.CheckInterval)

	defer onlineUserInquiry.Close()
	log.Println(client_announcer.RunHttpServer(cnf.HttpPort, onlineUserInquiry, chatConnRepo))
}

// TODO: Add tag for any application log
func initLogService(logConfig Log) {
	switch logConfig.LogLevel {
	case "info":
		log.SetLevel(log.InfoLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	default:
		log.SetLevel(log.WarnLevel)
	}

	if logConfig.Format == "json" {
		log.SetFormatter(&log.JSONFormatter{})
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
