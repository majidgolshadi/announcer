package client_announcer

import (
	log "github.com/sirupsen/logrus"
)

type onlineUserInquiry struct {
	mysqlConn *Mysql
	redisConn *Redis
}

func OnlineUserInquiryFactory(mysqlAddress string, mysqlUsername string, mysqlPassword string, mysqlDatabase string,
	redisAddr string, redisPassword string, redisDb int, redisHashTable string, redisCheckInterval int) (ouq *onlineUserInquiry, err error) {

	ouq = &onlineUserInquiry{}
	ouq.mysqlConn, err = mysqlClientFactory(mysqlAddress, mysqlUsername, mysqlPassword, mysqlDatabase)

	if err != nil {
		return nil, err
	}

	ouq.redisConn = &Redis{
		Address:       redisAddr,
		Password:      redisPassword,
		Database:      redisDb,
		HashTable:     redisPassword,
		CheckInterval: redisCheckInterval,
	}
	ouq.redisConn.connect()

	return ouq, nil
}

func (ouq *onlineUserInquiry) GetOnlineUsers(channel int) ([]string, error) {
	if channel < 0 {
		return ouq.redisConn.getAllUsers()
	}

	users, err := ouq.mysqlConn.getChannelUsers(channel)
	if err != nil {
		return nil, err
	}

	var onlineUsers []string
	for users.Next() {
		var username string
		if err := users.Scan(&username); err != nil {
			return nil, err
		}

		if ouq.redisConn.usernameExists(username) {
			onlineUsers = append(onlineUsers, username)
		}
	}

	return onlineUsers, nil
}

func (ouq *onlineUserInquiry) Close() {
	ouq.mysqlConn.close()
	log.Info("connection to mysql ", ouq.mysqlConn.address, " closed")

	ouq.redisConn.close()
	log.Info("connection to redis ", ouq.redisConn.Address, " closed")
}
