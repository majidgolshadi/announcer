package main

import (
	"github.com/BurntSushi/toml"
	"github.com/majidgolshadi/client-announcer"
	log "github.com/sirupsen/logrus"
	"strings"
)

type config struct {
	HttpPort string `toml:"rest_api_port"`

	Ejabberd  Ejabberd
	Client    Client
	Component Component
	Zookeeper Zookeeper
	Mysql     Mysql
	Redis     Redis
}

type Ejabberd struct {
	ClusterNodes   string `toml:"cluster_nodes"`
	DefaultCluster string `toml:"default_cluster"`
}

type Client struct {
	Username     string `toml:"username"`
	Password     string `toml:"password"`
	Domain       string `toml:"domain"`
	PingInterval int    `toml:"ping_interval"`
	RateLimit    int    `toml:"rate_limit"`
}

type Component struct {
	Name         string `toml:"name"`
	Secret       string `toml:"secret"`
	PingInterval int    `toml:"ping_interval"`
	RateLimit    int    `toml:"rate_limit"`
}

type Zookeeper struct {
	ClusterNodes string `toml:"cluster_nodes"`
	NameSpace    string `toml:"namespace"`
}

type Redis struct {
	ClusterNodes  string `toml:"cluster_nodes"`
	Password      string `toml:"password"`
	DB            int    `toml:"db"`
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

	ejabberdNodeAdd := strings.Split(cnf.Ejabberd.ClusterNodes, ",")
	if cnf.Component.Secret != "" {
		cluster, err = client_announcer.ClusterComponentFactory(
			ejabberdNodeAdd,
			cnf.Component.Name,
			cnf.Component.Secret,
			cnf.Component.PingInterval,
			cnf.Component.RateLimit)

	} else if cnf.Client.Password != "" {
		cluster, err = client_announcer.ClusterClientFactory(
			ejabberdNodeAdd,
			cnf.Client.Username,
			cnf.Client.Password,
			cnf.Client.Domain,
			cnf.Client.PingInterval,
			cnf.Client.RateLimit)
	}

	if err != nil {
		log.Fatal("erjaberd create connection error ", err.Error())
		return
	}

	err = chatConnRepo.Save(cnf.Ejabberd.DefaultCluster, cluster)
	if err != nil {
		log.Fatal("store in repository error ", err.Error())
		return
	}

	onlineUserInquiry, _ := client_announcer.OnlineUserInquiryFactory(
		cnf.Mysql.Address, cnf.Mysql.Username, cnf.Mysql.Password, cnf.Mysql.DB,
		cnf.Redis.ClusterNodes, cnf.Redis.Password, cnf.Redis.DB, cnf.Redis.CheckInterval)

	defer onlineUserInquiry.Close()

	client_announcer.RunHttpServer(cnf.HttpPort, onlineUserInquiry, chatConnRepo)
}
