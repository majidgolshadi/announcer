package client_announcer

import (
	"github.com/go-redis/redis"
)

type onlineUserInquiry struct {
	MysqlInquiry *MysqlInquiry
	redisConn    *redis.Client
}

func OnlineUserInquiryFactory(mysqlInquiry *MysqlInquiry, redisStr *Redis) (ouq *onlineUserInquiry, err error) {
	ouq = &onlineUserInquiry{}
	ouq.MysqlInquiry = mysqlInquiry
	ouq.MysqlInquiry.Connect()

	if err != nil {
		return nil, err
	}

	ouq.redisConn = redis.NewClient(&redis.Options{
		Addr:     redisStr.Address,
		Password: redisStr.Password,
		DB:       redisStr.Database,
	})

	ouq.redisConn.Ping()

	return
}

func (ouq *onlineUserInquiry) GetOnlineUsers(channel int) ([]string, error) {
	users, err := ouq.MysqlInquiry.getChannelUsers(channel)
	if err != nil {
		return nil, err
	}

	var onlineUsers []string
	for users.Next() {
		var username string
		if err := users.Scan(&username); err != nil {
			return nil, err
		}

		value, _ := ouq.redisConn.Get(username).Result()
		if value != "" {
			onlineUsers = append(onlineUsers, username)
		}
	}

	return onlineUsers, nil
}

func (ouq *onlineUserInquiry) Close() {
	ouq.MysqlInquiry.Close()
	ouq.redisConn.Close()
}
