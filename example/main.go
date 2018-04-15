package main

import (
	"github.com/BurntSushi/toml"
	"github.com/majidgolshadi/client-announcer"
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
	ClusterNodes string `toml:"cluster_nodes"`
}

type Client struct {
	Username     string `toml:"username"`
	Password     string `toml:"password"`
	Domain       string `toml:"domain"`
	PingInterval int    `toml:"ping_interval"`
}

type Component struct {
	Username string `toml:"username"`
	Secret   string `toml:"secret"`
	Name     string `toml:"name"`
}

type Zookeeper struct {
	ClusterNodes string `toml:"cluster_nodes"`
	NameSpace    string `toml:"namespace"`
}

type Redis struct {
	ClusterNodes string `toml:"cluster_nodes"`
	Password     string `toml:"password"`
	DB           int    `toml:"db"`
}

type Mysql struct {
	Address  string `toml:"address"`
	Username string `toml:"username"`
	Password string `toml:"password"`
	DB       string `toml:"db"`
}

func main() {
	var (
		cnf config
		cluster *client_announcer.Cluster
	)

	if _, err := toml.DecodeFile("config.toml", &cnf); err != nil {
		println(err.Error())
		return
	}

	chatConnRepo, err := client_announcer.ChatServerConnRepositoryFactory(cnf.Zookeeper.ClusterNodes, cnf.Zookeeper.NameSpace)
	if err != nil {
		println(err.Error())
		return
	}

	defer chatConnRepo.Close()

	if cnf.Component.Secret != "" {
		cluster, err = client_announcer.ClusterComponentFactory(
			cnf.Component.Name, cnf.Component.Username, cnf.Component.Secret)

	} else if cnf.Client.Password != "" {
		cluster, err = client_announcer.ClusterClientFactory(
			cnf.Client.Username, cnf.Client.Password, cnf.Client.Domain,
			cnf.Client.PingInterval, cnf.Ejabberd.ClusterNodes)
	}

	if err != nil {
		println(err.Error())
		return
	}

	chatConnRepo.SetCluster("A", cluster)

	onlineUserInquiry, _ := client_announcer.OnlineUserInquiryFactory(&client_announcer.MysqlInquiry{
		Address:  cnf.Mysql.Address,
		Username: cnf.Mysql.Username,
		Password: cnf.Mysql.Password,
		Database: cnf.Mysql.DB,
	}, &client_announcer.Redis{
		Address:  cnf.Redis.ClusterNodes,
		Password: cnf.Redis.Password,
		Database: cnf.Redis.DB,
	})

	defer onlineUserInquiry.Close()

	client_announcer.RunHttpServer(cnf.HttpPort, onlineUserInquiry, chatConnRepo)
}
