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

	monitRedisTime            float64
	monitChannelUserNum       int
	monitChannelOnlineUserNum int
}

const SoroushChannelId = "officialsoroushchannel"

func (ca *ChannelActor) Listen(chanAct <-chan *ChannelAct, msgChan chan<- *output.Msg) error {
	var start time.Time

	for rec := range chanAct {
		start = time.Now()
		if err := ca.sentToOnlineUser(rec.ChannelID, rec.MessageTemplate, msgChan); err != nil {
			log.Error("channel actor fetch channel ", rec.ChannelID, " online users error: ", err.Error())
		}

		log.WithFields(log.Fields{
			"redis": ca.monitRedisTime,
			"mysql": time.Now().Sub(start).Seconds(),
		}).Debug("channel actor process time consumption")

		ca.monitRedisTime = 0
		ca.monitChannelUserNum = 0
		ca.monitChannelOnlineUserNum = 0
	}

	return nil
}

func (ca *ChannelActor) sentToOnlineUser(channelID string, template string, msgChan chan<- *output.Msg) error {
	// all users are soroush official channel members
	if channelID == SoroushChannelId {
		userChan, err := ca.UserActivity.GetAllOnlineUsers()
		if err != nil {
			return err
		}

		for user := range userChan {
			msgChan <- &output.Msg{
				Temp: template,
				User: user,
			}
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
			msgChan <- &output.Msg{
				Temp: template,
				User: username,
			}
		}
	}

	log.WithFields(log.Fields{
		"users":      ca.monitChannelUserNum,
		"online":     ca.monitChannelOnlineUserNum,
		"channel_id": channelID,
	}).Debug("channel actor")

	return nil
}

func (ca *ChannelActor) Close() {
	log.Warn("channel actor close")
	ca.UserActivity.Close()
	ca.ChannelDataStore.Close()
}
