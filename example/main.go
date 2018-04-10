package main

import (
	"github.com/BurntSushi/toml"
	"github.com/majidgolshadi/client-announcer"
)

type config struct {
	Username string
	Password string
	Domain string
	HttpPort string `toml:"rest_api_port"`
	ClientPingInterval int `toml:"client_ping_interval"`
	EjabberdClusterNodes string `toml:"ejabberd_cluster_nodes"`
	ZookeeperClusterNodes string `toml:"zookeeper_cluster_nodes"`
	ZookeeperNamespace string `toml:"zookeeper_namespace"`
}

func main() {
	var cnf config
	if _, err := toml.DecodeFile("config.toml", &cnf); err != nil {
		println(err.Error())
		return
	}

	repo, err := client_announcer.RepositoryFactory(cnf.ZookeeperClusterNodes, cnf.ZookeeperNamespace)
	if err != nil {
		println(err.Error())
		return
	}

	defer repo.Close()

	cluster, err := client_announcer.ClusterClientFactory(
		cnf.Username, cnf.Password, cnf.Domain, cnf.ClientPingInterval, cnf.EjabberdClusterNodes)

	repo.SetCluster("A", cluster)
	client_announcer.RunHttpServer(cnf.HttpPort)
}
