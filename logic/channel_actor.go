package logic

import (
	"database/sql"
	"fmt"
	"github.com/majidgolshadi/client-announcer/output"
	log "github.com/sirupsen/logrus"
	"time"
)

type ChannelAct struct {
	MessageTemplate string
	ChannelID       string
}

type ChannelActor struct {
	Redis          *redis
	RedisHashTable string
	Mysql          *mysql

	monitRedisTime            float64
	monitChannelUserNum       int
	monitChannelOnlineUserNum int
}

const SoroushChannelId = "officialsoroushchannel"
const ChannelIDQuery = `select channel_id from ws_channel_data where channel_channelid="%s"`
const UsersChannelUsernameQuery = `select member_username from ws_channel_members where member_channelid="%s"`

func (ca *ChannelActor) Listen(chanAct <-chan *ChannelAct, msgChan chan<- *output.Msg) error {
	var start time.Time
	ca.Redis.connectAndKeep()
	if err := ca.Mysql.connect(); err != nil {
		log.Error("mysql connection error: ", err.Error())
		return err
	}

	for rec := range chanAct {
		start = time.Now()
		if err := ca.sentToOnlineUser(rec.ChannelID, rec.MessageTemplate, msgChan); err != nil {
			log.WithField("error", err.Error()).Error("fetch channel ", rec.ChannelID, " online users")
		}

		log.WithFields(log.Fields{
			"redis": ca.monitRedisTime,
			"mysql": time.Now().Sub(start).Seconds(),
		}).Debug("process time consumption")

		ca.monitRedisTime = 0
		ca.monitChannelUserNum = 0
		ca.monitChannelOnlineUserNum = 0
	}

	return nil
}

func (ca *ChannelActor) sentToOnlineUser(channelID string, template string, msgChan chan<- *output.Msg) error {
	// all users are soroush official channel members
	if channelID == SoroushChannelId {
		users, _ := ca.Redis.HKeys(ca.RedisHashTable)
		for _, user := range users {
			msgChan <- &output.Msg{
				Temp: template,
				User: user,
			}
		}

		return nil
	}

	rowUsers, err := ca.getChannelUsers(channelID)
	if err != nil {
		return err
	}
	defer rowUsers.Close()

	// loop on channel username fetch from mysql
	var username string
	redisStart := time.Now()
	for rowUsers.Next() {
		ca.monitChannelUserNum++
		if err := rowUsers.Scan(&username); err != nil {
			return err
		}

		// Is he/she online
		if result, _ := ca.Redis.HGet(ca.RedisHashTable, username); result != "" {
			ca.monitChannelOnlineUserNum++
			msgChan <- &output.Msg{
				Temp: template,
				User: username,
			}
		}
	}

	ca.monitRedisTime = time.Now().Sub(redisStart).Seconds()
	log.WithFields(log.Fields{
		"users":      ca.monitChannelUserNum,
		"online":     ca.monitChannelOnlineUserNum,
		"channel_id": channelID,
	}).Debug("channel users")

	return nil
}

func (ca *ChannelActor) getChannelUsers(channelID string) (result *sql.Rows, err error) {
	var id string
	if err = ca.Mysql.GetConnection().QueryRow(fmt.Sprintf(ChannelIDQuery, channelID)).Scan(&id); err != nil {
		return nil, err
	}

	// unfortunately return method directly make result untrustable
	result, err = ca.Mysql.GetConnection().Query(fmt.Sprintf(UsersChannelUsernameQuery, id))
	return result, err
}

func (ca *ChannelActor) Close() {
	log.Warn("close channel actor")
	ca.Redis.close()
	ca.Mysql.close()
}
