package client_announcer

import (
	log "github.com/sirupsen/logrus"
	"database/sql"
	"fmt"
)

type onlineUserInquiry struct {
	mysqlConn *mysql
	redisConn *redisDs

	redisHashTable string
}

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

	err = ouq.redisConn.connectAlways()
	return ouq, err
}

func (ouq *onlineUserInquiry) GetOnlineUsers(channel int) ([]string, error) {
	if channel < 0 {
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

	log.WithFields(log.Fields{"channel users": channelUserCount, "online users": len(onlineUsers)})
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
func (ouq *onlineUserInquiry) getAllOnlineUsers() ([]string, error) {
	return ouq.redisConn.connection().HKeys(ouq.redisHashTable).Result()
}

// fetch from mysql
func (ouq *onlineUserInquiry) getChannelUsers(channelID int) (rows *sql.Rows, err error) {
	query := fmt.Sprintf("select member_username from ws_channel_members where member_channelid='%d'", channelID)
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
