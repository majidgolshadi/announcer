package client_announcer

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
)

type MysqlInquiry struct {
	Username string
	Password string
	Address  string
	Database string

	connection *sql.DB
}

func (ms *MysqlInquiry) Connect() (err error) {
	dataSourceName := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8", ms.Username, ms.Password, ms.Address, ms.Database)
	ms.connection, err = sql.Open("mysql", dataSourceName)
	ms.connection.Ping()
	return
}

func (ms *MysqlInquiry) getChannelUsers(channelId int) (*sql.Rows, error) {
	query := fmt.Sprintf("select member_username from ws_channel_members where member_channelid='%d'", channelId)
	return ms.connection.Query(query)
}

func (ms *MysqlInquiry) Close() {
	ms.connection.Close()
}
