package logic

import (
	log "github.com/sirupsen/logrus"
	"time"
	"fmt"
)

type ChannelAct struct {
	MessageTemplate string
	ChannelID       string
}

type ChannelActor struct {
	UserActivity     UserActivity
	ChannelDataStore ChannelDataStore
	Domain string

	monitChannelUserNum       int
	monitChannelOnlineUserNum int
}

const SoroushChannelId = "officialsoroushchannel"

func (ca *ChannelActor) Listen(chanAct <-chan *ChannelAct, msgChan chan<- string) error {
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

func (ca *ChannelActor) createMessage(template string, user string) string {
	return fmt.Sprintf(template, fmt.Sprintf("%s@%s", user, ca.Domain))
}

func (ca *ChannelActor) sentToOnlineUser(channelID string, template string, msgChan chan<- string) error {

	// all users are soroush official channel members
	if channelID == SoroushChannelId {
		userChan, err := ca.UserActivity.GetAllOnlineUsers()
		if err != nil {
			return err
		}

		for user := range userChan {
			msgChan <- ca.createMessage(template, user)
		}

		return nil
	}

	// loop on channel username fetch from mysql
	userChan, err := ca.ChannelDataStore.GetChannelMembers(channelID)
	if err != nil {
		return err
	}

	for username := range userChan {
		ca.monitChannelUserNum++
		// Is he/she online
		if ca.UserActivity.IsHeOnline(username) {
			ca.monitChannelOnlineUserNum++
			msgChan <- ca.createMessage(template, username)
		}
	}

	return nil
}

func (ca *ChannelActor) Close() {
	log.Warn("channel actor close")
	ca.UserActivity.Close()
	ca.ChannelDataStore.Close()
}
