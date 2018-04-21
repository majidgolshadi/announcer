package client_announcer

import (
	log "github.com/sirupsen/logrus"
	"os"
)

func init() {
	f, _ := os.Create("announcer.log")
	log.SetFormatter(&log.TextFormatter{})
	log.SetOutput(f)
}
