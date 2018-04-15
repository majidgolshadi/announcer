package client_announcer

import "github.com/go-redis/redis"

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

	if statusCmd := ouq.redisConn.Ping(); statusCmd.Err() != nil {
		return nil, statusCmd.Err()
	}

	return ouq, nil
}

func (ouq *onlineUserInquiry) GetOnlineUsers(channel int) (map[string]string, error) {
	users, err := ouq.MysqlInquiry.getChannelUsers(channel)
	if err != nil {
		return nil, err
	}

	var onlineUsers map[string]string
	for users.Next() {
		var username string
		if err := users.Scan(&username); err != nil {
			return nil, err
		}

		if value, _ := ouq.redisConn.Get(username).Result(); value != "" {
			onlineUsers[username] = value
		}
	}

	return onlineUsers, nil
}

func (ouq *onlineUserInquiry) Close() {
	ouq.MysqlInquiry.Close()
	ouq.redisConn.Close()
}
