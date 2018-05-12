package logic

import (
	"database/sql"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/majidgolshadi/client-announcer/output"
)

type ChannelAct struct {
	MessageTemplate string
	ChannelID       string
}

type ChannelActor struct {
	Redis          *redis
	RedisHashTable string
	Mysql          *mysql
}

const SoroushChannelId = "officialsoroushchannel"
const ChannelIDQuery = `select channel_id from ws_channel_data where channel_channelid="%s"`
const UsersChannelUsernameQuery = `select member_username from ws_channel_members where member_channelid="%s"`

func (ca *ChannelActor) Listen(chanAct <-chan *ChannelAct, msgChan chan<- *output.Msg) error {
	ca.Redis.connectAndKeep()
	if err := ca.Mysql.connect(); err != nil {
		log.WithField("error", err.Error()).Error("mysql connection error")
		return err
	}

	for rec := range chanAct {

		users, err := ca.getWhoSendTo(rec.ChannelID)
		if err != nil {
			log.WithField("error", err.Error()).Error("find user list error")
			continue
		}

		for _,user := range users {
			msgChan <- &output.Msg{
				Temp: rec.MessageTemplate,
				User: user,
			}
		}
	}

	return nil
}

func (ca *ChannelActor) getWhoSendTo(channelID string) (users []string, err error) {
	if channelID == SoroushChannelId {
		users, _ = ca.Redis.HKeys(ca.RedisHashTable)
		return users, nil
	}

	rowUsers, err := ca.getChannelUsers(channelID)
	if err != nil {
		return []string{}, err
	}
	defer rowUsers.Close()

	channelUserCount := 0
	var onlineUsers []string
	var username string
	for rowUsers.Next() {
		channelUserCount++
		if err := rowUsers.Scan(&username); err != nil {
			return nil, err
		}

		// Is he/she online
		if result, _ := ca.Redis.HGet(ca.RedisHashTable, username); result != "" {
			onlineUsers = append(onlineUsers, username)
		}
	}

	log.WithFields(log.Fields{
		"channel_users": channelUserCount,
		"online": len(onlineUsers),
		"channel_id": channelID}).
	Info("channel users")

	return onlineUsers, nil
}

func (ca *ChannelActor) getChannelUsers(channelID string) (result *sql.Rows, err error) {
	var id string
	err = ca.Mysql.GetConnection().QueryRow(fmt.Sprintf(ChannelIDQuery, channelID)).Scan(&id)
	if err != nil {
		return nil, err
	}

	result, err = ca.Mysql.GetConnection().Query(fmt.Sprintf(UsersChannelUsernameQuery, id))
	if err != nil {
		log.Warn("mysql result error ", err.Error())
	}

	return result, err
}

func (ca *ChannelActor) Close() {
	log.Warn("close channel actor")
	ca.Redis.close()
	ca.Mysql.close()
}
