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
	ClusterNodes string `toml:"cluster_nodes"`
}

func main() {
	var cnf config
	if _, err := toml.DecodeFile("config.toml", &cnf); err != nil {
		println(err.Error())
		return
	}

	cluster, err := client_announcer.ClusterFactory(cnf.Username, cnf.Password, cnf.Domain, 2, cnf.ClusterNodes )
	if err != nil {
		log.Fatal(err.Error())
	}

	for i := 0; i< 2; i++ {
		if err := cluster.Send("<message xml:lang='en' to='2@soroush.ir' from='4@soroush.ir/announcer' type='chat' id='E0V0Z-527' xmlns='jabber:client'><body xml:lang='SEND_TIME_IN_GMT'>1521973339450</body><composing xmlns='http://jabber.org/protocol/chatstates'/></message>", "2@soroush.ir"); err != nil {
		//if err := cluster.Send("2@soroush.ir", "<message xml:lang='en' to='2@soroush.ir' from='1@soroush.ir/announcer' type='chat' id='E0V0Z-527' xmlns='jabber:client'><body xml:lang='SEND_TIME_IN_GMT'>1521973339450</body><composing xmlns='http://jabber.org/protocol/chatstates'/></message>"); err != nil {
		//if err := cluster.Send("2@soroush.ir", "test message"); err != nil {
			println(err.Error())
		}
		println("done")
	}
}
