package logic

import (
	"github.com/majidgolshadi/client-announcer/output"
	log "github.com/sirupsen/logrus"
	"time"
)

type ChannelAct struct {
	MessageTemplate string
	ChannelID       string
}

type ChannelActor struct {
	UserActivity     UserActivity
	ChannelDataStore ChannelDataStore
	askBuffer        int
	buffer           []string

	monitChannelUserNum       int
	monitChannelOnlineUserNum int
}

func (ca *ChannelActor) Listen(askBuffer int, chanAct <-chan *ChannelAct, msgChan chan<- *output.Message) error {
	ca.askBuffer = askBuffer
	var start time.Time

	for rec := range chanAct {
		start = time.Now()
		if err := ca.sentToOnlineUser(rec.ChannelID, rec.MessageTemplate, msgChan); err != nil {
			log.Error("channel actor fetch channel ", rec.ChannelID, " online users error: ", err.Error())
		}

		log.WithFields(log.Fields{
			"users":        ca.monitChannelUserNum,
			"online":       ca.monitChannelOnlineUserNum,
			"channel_id":   rec.ChannelID,
			"process_time": time.Now().Sub(start).Seconds(),
		}).Info("channel actor")

		ca.monitChannelUserNum = 0
		ca.monitChannelOnlineUserNum = 0
	}

	return nil
}

func (ca *ChannelActor) sentToOnlineUser(channelID string, template string, msgChan chan<- *output.Message) error {
	userChan, err := ca.ChannelDataStore.GetChannelMembers(channelID)
	if err != nil {
		return err
	}

	for username := range userChan {
		ca.buffer = append(ca.buffer, username)

		if len(ca.buffer) >= ca.askBuffer {
			ca.bulkAskFromUserActivity(template, msgChan)
		}

		ca.monitChannelUserNum++
	}

	ca.bulkAskFromUserActivity(template, msgChan)

	return nil
}

func (ca *ChannelActor) bulkAskFromUserActivity(template string, msgChan chan<- *output.Message) {
	usernames := ca.UserActivity.FilterOnlineUsers(ca.buffer)

	for _, username := range usernames {
		ca.monitChannelOnlineUserNum++

		msgChan <- &output.Message{
			Template: template,
			Username: username,
			Loggable: false,
		}
	}

	// reset users array
	ca.buffer = nil
}

func (ca *ChannelActor) Close() {
	log.Warn("channel actor close")
	ca.UserActivity.Close()
	ca.ChannelDataStore.Close()
}
