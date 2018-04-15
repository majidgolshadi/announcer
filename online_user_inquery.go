package client_announcer

type onlineUserInquiry struct {
	MysqlConn *Mysql
	redisConn *Redis
}

func OnlineUserInquiryFactory(mysqlAddress string, mysqlUsername string, mysqlPassword string, mysqlDatabase string,
	redisAddr string, redisPassword string, redisDb int, redisPingInterval int) (ouq *onlineUserInquiry, err error) {

	ouq = &onlineUserInquiry{}
	ouq.MysqlConn, err = MysqlClientFactory(mysqlAddress, mysqlUsername, mysqlPassword, mysqlDatabase)

	if err != nil {
		return nil, err
	}

	ouq.redisConn = RedisClientFactory(redisAddr, redisPassword, redisDb, redisPingInterval)

	return ouq, nil
}

func (ouq *onlineUserInquiry) GetOnlineUsers(channel int) ([]string, error) {
	users, err := ouq.MysqlConn.GetChannelUsers(channel)
	if err != nil {
		return nil, err
	}

	var onlineUsers []string
	for users.Next() {
		var username string
		if err := users.Scan(&username); err != nil {
			return nil, err
		}

		if ouq.redisConn.UsernameExists(username) {
			onlineUsers = append(onlineUsers, username)
		}
	}

	return onlineUsers, nil
}

func (ouq *onlineUserInquiry) Close() {
	ouq.MysqlConn.Close()
	ouq.redisConn.Close()
}
