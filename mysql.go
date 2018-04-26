package client_announcer

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	log "github.com/sirupsen/logrus"
)

type mysql struct {
	Username string
	Password string
	Address  string
	Database string
	Charset  string

	connection *sql.DB
}

func (ms *mysql) connect() (err error) {
	if ms.Charset != "" {
		ms.Charset = "utf8"
	}

	dataSourceName := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=%s",
		ms.Username, ms.Password, ms.Address, ms.Database, ms.Charset)

	if ms.connection, err = sql.Open("mysql", dataSourceName); err != nil {
		return err
	}

	return ms.connection.Ping()
}

func (ms *mysql) close() {
	log.Info("close mysql connections to ", ms.Address)
	ms.connection.Close()
}
