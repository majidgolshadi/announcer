package client_announcer

import (
	"fmt"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

type Mysql struct {
	username string
	password string
	address  string
	database string

	connection *sql.DB
}

func mysqlClientFactory(address string, username string, password string, database string) (*Mysql, error) {
	m := &Mysql{
		address:  address,
		username: username,
		password: password,
		database: database,
	}

	if err := m.connect(); err != nil {
		return nil, err
	}

	return m, nil
}

func (ms *Mysql) connect() (err error) {
	dataSourceName := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8", ms.username, ms.password, ms.address, ms.database)
	ms.connection, err = sql.Open("mysql", dataSourceName)
	ms.connection.Ping()
	return
}

func (ms *Mysql) getChannelUsers(channelId int) (*sql.Rows, error) {
	query := fmt.Sprintf("select member_username from ws_channel_members where member_channelid='%d'", channelId)
	return ms.connection.Query(query)
}

func (ms *Mysql) close() {
	ms.connection.Close()
}
