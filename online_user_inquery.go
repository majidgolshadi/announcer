package client_announcer

import (
	"database/sql"
	"fmt"
	log "github.com/sirupsen/logrus"
)

type onlineUserInquiry struct {
	mysql *mysql
	redis *redisDs

	redisHashTable string
}

const SoroushChannelId = "officialsoroushchannel"
const UsersChannelUsernameQuery = "select member_username from ws_channel_members as `wm` INNER JOIN ws_channel_data as `wd` ON (wm.member_channelid = wd.channel_id) where wd.channel_channelid='%s'"

func OnlineUserInquiryFactory(mysqlAddress string, mysqlUsername string, mysqlPassword string, mysqlDatabase string,
	redisAddr string, redisPassword string, redisDb int, redisHashTable string, redisCheckInterval int) (ouq *onlineUserInquiry, err error) {

	ouq = &onlineUserInquiry{
		redisHashTable: redisHashTable,

		mysql: &mysql{
			Address:  mysqlAddress,
			Username: mysqlUsername,
			Password: mysqlPassword,
			Database: mysqlDatabase,
		},

		redis: &redisDs{
			Address:       redisAddr,
			Password:      redisPassword,
			Database:      redisDb,
			CheckInterval: redisCheckInterval,
		},
	}

	err = ouq.mysql.connect()
	if err != nil {
		return nil, err
	}

	// we only want to force connect to mysql only
	ouq.redis.retryToConnect()
	return ouq, err
}

func (ouq *onlineUserInquiry) GetOnlineUsers(channel string) ([]string, error) {
	if channel == SoroushChannelId {
		return ouq.getAllOnlineUsers()
	}

	users, err := ouq.getChannelUsers(channel)
	if err != nil {
		return nil, err
	}

	channelUserCount := 0
	var onlineUsers []string
	var username string
	for users.Next() {
		channelUserCount++
		if err := users.Scan(&username); err != nil {
			return nil, err
		}

		if ouq.IsOnline(username) {
			onlineUsers = append(onlineUsers, username)
		}
	}

	log.WithFields(log.Fields{"all": channelUserCount, "online": len(onlineUsers)}).Info("channel users")
	return onlineUsers, nil
}

// Fetch data from redis and if there is no connection to that it will be say online
func (ouq *onlineUserInquiry) IsOnline(username string) bool {
	if ouq.redis.connectionStatus() {
		result, _ := ouq.redis.connection().HGet(ouq.redisHashTable, username).Result()
		return result != ""
	}

	return true
}

// Get all online users from redis
func (ouq *onlineUserInquiry) getAllOnlineUsers() (users []string, err error) {
	if ouq.redis.connStatus {
		return ouq.redis.connection().HKeys(ouq.redisHashTable).Result()
	}

	return
}

// fetch from mysql
func (ouq *onlineUserInquiry) getChannelUsers(channelID string) (rows *sql.Rows, err error) {
	query := fmt.Sprintf(UsersChannelUsernameQuery, channelID)
	result, err := ouq.mysql.connection.Query(query)

	if err != nil {
		log.Warn("mysql result error ", err.Error())
	}

	return result, err
}

func (ouq *onlineUserInquiry) Close() {
	log.Info("close online user inquiry connections")
	ouq.mysql.close()
	ouq.redis.close()
}
