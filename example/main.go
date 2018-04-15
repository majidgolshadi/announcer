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
	ClusterNodes   string `toml:"cluster_nodes"`
	DefaultCluster string `toml:"default_cluster"`
}

type Client struct {
	Username     string `toml:"username"`
	Password     string `toml:"password"`
	Domain       string `toml:"domain"`
	PingInterval int    `toml:"ping_interval"`
}

type Component struct {
	Name   string `toml:"name"`
	Secret string `toml:"secret"`
}

type Zookeeper struct {
	ClusterNodes string `toml:"cluster_nodes"`
	NameSpace    string `toml:"namespace"`
}

type Redis struct {
	ClusterNodes string `toml:"cluster_nodes"`
	Password     string `toml:"password"`
	DB           int    `toml:"db"`
	PingInterval int    `toml:"ping_interval"`
}

type Mysql struct {
	Address  string `toml:"address"`
	Username string `toml:"username"`
	Password string `toml:"password"`
	DB       string `toml:"db"`
}

func main() {
	var (
		cnf     config
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
			cnf.Component.Name, cnf.Component.Secret, cnf.Ejabberd.ClusterNodes)

	} else if cnf.Client.Password != "" {
		cluster, err = client_announcer.ClusterClientFactory(
			cnf.Client.Username, cnf.Client.Password, cnf.Client.Domain,
			cnf.Client.PingInterval, cnf.Ejabberd.ClusterNodes)
	}

	if err != nil {
		println("erjaberd connection error:", err.Error())
		return
	}

	chatConnRepo.SetCluster(cnf.Ejabberd.DefaultCluster, cluster)
	chatConnRepo.SetDefaultCluster(cnf.Ejabberd.DefaultCluster)

	onlineUserInquiry, _ := client_announcer.OnlineUserInquiryFactory(
		cnf.Mysql.Address, cnf.Mysql.Username, cnf.Mysql.Password, cnf.Mysql.DB,
		cnf.Redis.ClusterNodes, cnf.Redis.Password, cnf.Redis.DB, cnf.Redis.PingInterval)

	defer onlineUserInquiry.Close()

	client_announcer.RunHttpServer(cnf.HttpPort, onlineUserInquiry, chatConnRepo)
}
