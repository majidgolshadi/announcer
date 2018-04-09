package main

import (
	"github.com/BurntSushi/toml"
	"github.com/majidgolshadi/client-announcer"
	"log"
)

type config struct {
	Username string
	Password string
	Domain string
	HttpPort string `toml:"rest_api_port"`
	ClientPingInterval int `toml:"client_ping_interval"`
	ClusterNodes string `toml:"cluster_nodes"`
}

func main() {
	var cnf config
	if _, err := toml.DecodeFile("config.toml", &cnf); err != nil {
		println(err.Error())
		return
	}

	client_announcer.RunHttpServer(cnf.HttpPort)

	cluster, err := client_announcer.ClusterFactory(
		cnf.Username, cnf.Password, cnf.Domain,
		cnf.ClientPingInterval, cnf.ClusterNodes)

	if err != nil {
		log.Fatal(err.Error())
	}

	cluster.Send("%smsg", "majid")
}
