package client_announcer

import (
	"database/sql"
	"fmt"
	log "github.com/sirupsen/logrus"
)

type onlineUserInquiry struct {
	mysqlConn *mysql
	redisConn *redisDs

	redisHashTable string
}

const SoroushChannelId  = "officialsoroushchannel"

func OnlineUserInquiryFactory(mysqlAddress string, mysqlUsername string, mysqlPassword string, mysqlDatabase string,
	redisAddr string, redisPassword string, redisDb int, redisHashTable string, redisCheckInterval int) (ouq *onlineUserInquiry, err error) {

	ouq = &onlineUserInquiry{
		redisHashTable: redisHashTable,

		mysqlConn: &mysql{
			Address:  mysqlAddress,
			Username: mysqlUsername,
			Password: mysqlPassword,
			Database: mysqlDatabase,
		},

		redisConn: &redisDs{
			Address:       redisAddr,
			Password:      redisPassword,
			Database:      redisDb,
			CheckInterval: redisCheckInterval,
		},
	}

	err = ouq.mysqlConn.connect()
	if err != nil {
		return nil, err
	}

	// we only want to force connect to mysql only
	ouq.redisConn.retryToConnect()
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
	if ouq.redisConn.connectionStatus() {
		result, _ := ouq.redisConn.connection().HGet(ouq.redisHashTable, username).Result()
		return result != ""
	}

	return true
}

// Get all online users from redis
func (ouq *onlineUserInquiry) getAllOnlineUsers() (users []string, err error) {
	if ouq.redisConn.connStatus {
		return ouq.redisConn.connection().HKeys(ouq.redisHashTable).Result()
	}

	return
}

// fetch from mysql
func (ouq *onlineUserInquiry) getChannelUsers(channelID string) (rows *sql.Rows, err error) {
	query := fmt.Sprintf("select member_username from ws_channel_members as `wm` INNER JOIN ws_channel_data as `wd` ON (wm.member_channelid = wd.channel_id) where wd.channel_channelid='%s'", channelID)
	result, err := ouq.mysqlConn.connection.Query(query)

	if err != nil {
		log.Warn("mysql result error ", err.Error())
	}

	return result, err
}

func (ouq *onlineUserInquiry) Close() {
	log.Info("close online user inquiry connections")
	ouq.mysqlConn.close()
	ouq.redisConn.close()
}
